package gst_audiomixer

import (
	"context"
	"fmt"
	"github.com/notedit/gst"
	"log"
)



func StartServer(ctx context.Context, portStart, portEnd int, sampleChan chan []byte) error {
	if portEnd < portStart {
		return fmt.Errorf("[AUDIOMIXER] invalid ports")
	}

	pipelineString := ""
	for i := portStart; i <= portEnd; i++ {
		pipelineString += fmt.Sprintf("udpsrc port=%d caps=\"application/x-rtp,media=audio,encoding-name=L16,clock-rate=48000,channels=2\" ! rtpL16depay ! adder.\n", i)
	}
	pipelineString += "liveadder name=adder ! opusenc ! appsink name=sink"

	log.Print(pipelineString)
	pipeline, err := gst.ParseLaunch(pipelineString)
	if err != nil {
		return fmt.Errorf("[AUDIOMIXER] error parsing pipeline")
	}

	pipeline.SetState(gst.StatePlaying)
	sinkEl := pipeline.GetByName("sink")
	if sinkEl == nil {
		return fmt.Errorf("[AUDIOMIXER] sinkEl not found")
	}

	go func() {
		for {
			sample, _ := sinkEl.PullSample()
			if sample != nil {
				sampleChan <- sample.Data
			}
		}
	}()

	select {
	case <-ctx.Done():
		close(sampleChan)
		return nil
	}
}