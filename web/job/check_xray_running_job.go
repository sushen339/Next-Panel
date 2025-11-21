package job

import (
	"x-ui/logger"
	"x-ui/web/service"
)

type CheckXrayRunningJob struct {
	xrayService service.XrayService

	checkTime int
}

func NewCheckXrayRunningJob() *CheckXrayRunningJob {
	return new(CheckXrayRunningJob)
}

func (j *CheckXrayRunningJob) Run() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("CheckXrayRunningJob panic recovered:", r)
		}
	}()

	if !j.xrayService.DidXrayCrash() {
		if j.checkTime > 0 {
			logger.Debug("CheckXrayRunningJob: Xray is now running normally")
		}
		j.checkTime = 0
	} else {
		j.checkTime++
		logger.Warningf("CheckXrayRunningJob: Xray crash detected (count: %d)", j.checkTime)
		
		// only restart if it's down 2 times in a row
		if j.checkTime > 1 {
			logger.Info("CheckXrayRunningJob: Attempting to restart Xray after multiple crash detections")
			err := j.xrayService.RestartXray(false)
			j.checkTime = 0
			if err != nil {
				logger.Error("CheckXrayRunningJob: Failed to restart Xray:", err)
			} else {
				logger.Info("CheckXrayRunningJob: Successfully restarted Xray")
			}
		}
	}
}
