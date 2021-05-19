package main

import (
	"fmt"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"time"
)

type StreamClient struct {
	filePath string
	audioTrack *webrtc.TrackLocalStaticSample
	logMode int
	bytesLoaded int
	bytesSpeed int
	cache []byte
	readable *ReadableClient
	finishedLoading bool
	finished bool
	bitsPerSample int
	sampleRate int
	channelCount int
	bufferLength int
	finishedBytes bool
	timePulseBuffer float64
	runningPulse bool
	lastLag int
	equalCount int
	lastBytesLoaded int
	lastByteCheck int
	lastByte int
	stopped bool
	paused bool
}

func Stream(
	filePath string,
	bitsPerSample int,
	sampleRate int,
	channelCount int,
	logMode int,
	bufferLength int,
) *StreamClient {
	var timePulseBuffer float64
	if bufferLength == 4{
		timePulseBuffer = 1.5
	}else{
		timePulseBuffer = 0
	}
	stream := &StreamClient{
		filePath: filePath,
		logMode: logMode,
		bitsPerSample: bitsPerSample,
		sampleRate: sampleRate,
		channelCount: channelCount,
		bufferLength: bufferLength,
		timePulseBuffer: timePulseBuffer,
		finished: true,
	}
	return stream
}
func (s *StreamClient) start(audioTrack *webrtc.TrackLocalStaticSample)  {
	s.audioTrack = audioTrack
	s.setReadable(s.filePath)
	go func() {
		s.processData()
	}()
}
func (s *StreamClient) setReadable(filePath string) bool {
	s.bytesLoaded = 0
	s.bytesSpeed = 0
	s.lastLag = 0
	s.equalCount = 0
	s.lastBytesLoaded = 0
	s.finishedBytes = false
	s.lastByteCheck = 0
	s.lastByte = 0
	s.readable = Readable(filePath)
	if s.stopped {
		return false
	}
	if s.readable != nil {
		s.finished = false
		s.finishedLoading = false
		s.readable.onData(func(bytes []byte) {
			s.bytesLoaded += len(bytes)
			s.bytesSpeed = len(bytes)
			if !s.needsBuffering(true) {
				s.readable.pause()
				s.runningPulse = false
				if s.logMode > 1 {
					fmt.Println(fmt.Sprintf("ENDED_BUFFERING -> %d", time.Now().Unix()))
					fmt.Println(fmt.Sprintf("BYTES_STREAM_CACHE_LENGTH -> %d", len(s.cache)))
					fmt.Println(fmt.Sprintf("PULSE -> %t", s.runningPulse))
				}
			}
			if s.logMode > 1 {
				fmt.Println(fmt.Sprintf("BYTES_LOADED -> %d OF %d", s.bytesLoaded, s.readable.getFilesizeInBytes()))
			}
			s.cache = append(s.cache, bytes[:]...)
		})
		s.readable.onEnd(func() {
			s.finishedLoading = true
			if s.logMode > 1 {
				fmt.Println(fmt.Sprintf("COMPLETED_BUFFERING -> %d", time.Now().Unix()))
				fmt.Println(fmt.Sprintf("BYTES_STREAM_CACHE_LENGTH -> %d", len(s.cache)))
				fmt.Println(fmt.Sprintf("BYTES_LOADED -> %d OF %d", s.bytesLoaded, s.readable.getFilesizeInBytes()))
			}
		})
		return true
	}else{
		return false
	}
}
func (s *StreamClient) needsBuffering(withPulseCheck bool) bool {
	if s.finishedLoading {
		return false
	}
	byteLength := ((s.sampleRate * s.bitsPerSample)/8/100) * s.channelCount
	result := len(s.cache) < (byteLength * 100 * s.bufferLength)
	result = result && (s.bytesLoaded < s.readable.getFilesizeInBytes() - s.bytesSpeed * 2 || s.finishedBytes)
	if s.timePulseBuffer > 0 && withPulseCheck {
		result = result && s.runningPulse
	}
	return result
}

func (s *StreamClient) checkLag() bool {
	if s.finishedLoading{
		return false
	}
	byteLength := ((s.sampleRate * s.bitsPerSample)/8/100) * s.channelCount
	return len(s.cache) < byteLength * 100
}
func (s *StreamClient) finish() {
	s.finished = true
}
func (s *StreamClient) processData(){
	oldTime := int(time.Now().Unix())
	if s.stopped{
		return
	}
	byteLength := ((s.sampleRate * s.bitsPerSample)/8/100) * s.channelCount
	if !(!s.finished && s.finishedLoading && len(s.cache) < byteLength){
		if s.needsBuffering(false){
			checkBuff := true
			if s.timePulseBuffer > 0 {
				s.runningPulse = float64(len(s.cache)) < float64(byteLength) * 100 * s.timePulseBuffer
				checkBuff = s.runningPulse
			}
			if s.readable != nil && checkBuff {
				if s.logMode > 1 {
					fmt.Println(fmt.Sprintf("PULSE -> %t", s.runningPulse))
				}
				s.readable.resume()
				if s.logMode > 1 {
					fmt.Println(fmt.Sprintf("BUFFERING -> %d", oldTime))
				}
			}
		}
		var fileSize int
		checkLag := s.checkLag()
		if oldTime - s.lastByteCheck > 1000 {
			fileSize = s.readable.getFilesizeInBytes()
			s.lastByte = fileSize
			s.lastByteCheck = oldTime
		}else{
			fileSize = s.lastByte
		}
		if !s.paused && !s.finished && (len(s.cache) >= byteLength || s.finishedLoading) && !checkLag{
			buffer := s.cache[:byteLength]
			s.cache = s.cache[byteLength:]
			//NOT WORKING THIS LINE
			err := s.audioTrack.WriteSample(media.Sample{
				Data:buffer,
			})
			if err != nil{
				fmt.Println(fmt.Sprintf("ERROR_WRITING_STREAM -> %s", err))
				return
			}
		}else if checkLag && s.logMode > 1 {
			fmt.Println(fmt.Sprintf("STREAM_LAG -> %d", oldTime))
			fmt.Println(fmt.Sprintf("BYTES_STREAM_CACHE_LENGTH -> %d", len(s.cache)))
			fmt.Println(fmt.Sprintf("BYTES_LOADED -> %d OF %d", s.bytesLoaded, s.readable.getFilesizeInBytes()))
		}
		if !s.finishedLoading {
			if fileSize == s.lastBytesLoaded{
				if s.equalCount >= 15 {
					s.equalCount = 0
					if s.logMode > 1 {
						fmt.Println(fmt.Sprintf("NOT_ENOUGH_BYTES -> %d", oldTime))
					}
					s.finishedBytes = true
					s.readable.resume()
				}else if oldTime - s.lastLag > 1000{
					s.equalCount += 1
					s.lastLag = oldTime
				}
			}else{
				s.lastBytesLoaded = fileSize
				s.equalCount = 0
				s.finishedBytes = false
			}
		}
	}
	if !s.finished && s.finishedLoading && len(s.cache) < byteLength{
		s.finish()
	}
	toSubtract := int(time.Now().Unix()) - oldTime
	var timeWait int
	if s.finished || s.paused || s.checkLag(){
		timeWait = 500
	}else{
		timeWait = 10
	}
	timeWait -= toSubtract
	time.Sleep(time.Millisecond * time.Duration(timeWait))
	s.processData()
}