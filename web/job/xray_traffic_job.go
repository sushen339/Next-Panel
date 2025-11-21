package job

import (
	"encoding/json"
	"fmt"
	"x-ui/logger"
	"x-ui/web/service"
	"x-ui/xray"

	"github.com/valyala/fasthttp"
)

type XrayTrafficJob struct {
	settingService  service.SettingService
	xrayService     service.XrayService
	inboundService  service.InboundService
	outboundService service.OutboundService
}

func NewXrayTrafficJob() *XrayTrafficJob {
	return new(XrayTrafficJob)
}

func (j *XrayTrafficJob) Run() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("XrayTrafficJob panic recovered:", r)
		}
	}()

	if !j.xrayService.IsXrayRunning() {
		logger.Debug("XrayTrafficJob: Xray is not running, skipping traffic collection")
		return
	}

	traffics, clientTraffics, err := j.xrayService.GetXrayTraffic()
	if err != nil {
		logger.Error("XrayTrafficJob: Failed to get Xray traffic:", err)
		return
	}
	logger.Debugf("XrayTrafficJob: Collected traffic for %d inbounds and %d clients", len(traffics), len(clientTraffics))

	err, needRestart0 := j.inboundService.AddTraffic(traffics, clientTraffics)
	if err != nil {
		logger.Error("XrayTrafficJob: Failed to add inbound traffic:", err)
	}

	err, needRestart1 := j.outboundService.AddTraffic(traffics, clientTraffics)
	if err != nil {
		logger.Error("XrayTrafficJob: Failed to add outbound traffic:", err)
	}

	ExternalTrafficInformEnable, err := j.settingService.GetExternalTrafficInformEnable()
	if err != nil {
		logger.Warning("XrayTrafficJob: Failed to get ExternalTrafficInformEnable setting:", err)
	} else if ExternalTrafficInformEnable {
		if err := j.informTrafficToExternalAPI(traffics, clientTraffics); err != nil {
			logger.Error("XrayTrafficJob: Failed to inform external API:", err)
		}
	}

	if needRestart0 || needRestart1 {
		logger.Info("XrayTrafficJob: Marking Xray for restart due to traffic changes")
		j.xrayService.SetToNeedRestart()
	}
}

func (j *XrayTrafficJob) informTrafficToExternalAPI(inboundTraffics []*xray.Traffic, clientTraffics []*xray.ClientTraffic) error {
	informURL, err := j.settingService.GetExternalTrafficInformURI()
	if err != nil {
		return fmt.Errorf("failed to get external traffic inform URI: %w", err)
	}

	if informURL == "" {
		return fmt.Errorf("external traffic inform URI is empty")
	}

	requestBody, err := json.Marshal(map[string]any{"clientTraffics": clientTraffics, "inboundTraffics": inboundTraffics})
	if err != nil {
		return fmt.Errorf("failed to marshal traffic data: %w", err)
	}

	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	request.Header.SetMethod("POST")
	request.Header.SetContentType("application/json; charset=UTF-8")
	request.SetBody(requestBody)
	request.SetRequestURI(informURL)

	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)

	if err := fasthttp.Do(request, response); err != nil {
		return fmt.Errorf("failed to POST to external API: %w", err)
	}

	statusCode := response.StatusCode()
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("external API returned non-success status code: %d", statusCode)
	}

	logger.Debug("Successfully informed external API about traffic")
	return nil
}
