package main

import (
	"encoding/json"
	"fmt"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
	"strings"
	"time"
)

type RTCPeerConnectionClient struct{
	rtcConnection *webrtc.PeerConnection
	chatId int
	inviteHash *string
	audioTrack *webrtc.TrackLocalStaticSample
}

func RTCPeerConnection(chatId int, inviteHash *string) *RTCPeerConnectionClient {
	resultClient := &RTCPeerConnectionClient{
		chatId: chatId,
		inviteHash: inviteHash,
	}
	return resultClient
}
func (r *RTCPeerConnectionClient) joinCall() bool{
	mediaEngine := webrtc.MediaEngine{}
	opusParams, err := opus.NewParams()
	if err != nil {
		panic(err)
	}
	/**offer := webrtc.SessionDescription{}
	Decode(MustReadStdin(), &offer)
	fmt.Println(offer.SDP)**/
	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithAudioEncoders(&opusParams),
	)
	codecSelector.Populate(&mediaEngine)
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err == nil{
		r.rtcConnection = peerConnection
		peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
			if logMode > 0 {
				fmt.Printf("IceConnection State has changed to %s \n", connectionState.String())
			}
		})
		stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
			Audio: func(constraints *mediadevices.MediaTrackConstraints) {
				constraints.SampleSize = prop.Int(2)
				constraints.ChannelCount = prop.Int(1)
				constraints.SampleRate = prop.Int(48000)
			},
			Codec: codecSelector,
		})
		if err != nil {
			fmt.Println(err)
			return false
		}
		for _, track := range stream.GetTracks() {
			track.OnEnded(func(err error) {
				fmt.Printf("Track (ID: %s) ended with error: %v\n",
					track.ID(), err)
			})
			fmt.Println("Found track",track.ID())
			_, err = peerConnection.AddTransceiverFromTrack(track,
				webrtc.RTPTransceiverInit{
					Direction: webrtc.RTPTransceiverDirectionSendonly,
				},
			)
			if err != nil {
				panic(err)
			}
		}
		offer, err := peerConnection.CreateOffer(nil)
		if err != nil {
			_ = peerConnection.Close()
			fmt.Println(err)
			return false
		}
		err = peerConnection.SetLocalDescription(offer)

		sdp := r.parseSdp(offer.SDP)
		if sdp.ufrag == nil || sdp.pwd == nil || sdp.hash == nil || sdp.fingerprint == nil || sdp.source == nil {
			return false
		}
		oggFile, err := oggwriter.New("output.ogg", 48000, 2)
		payload := JoinVoiceCallParams{
			chatId: r.chatId,
			ufrag: *sdp.ufrag,
			pwd: *sdp.pwd,
			hash: *sdp.hash,
			setup: "active",
			fingerprint: *sdp.fingerprint,
			source: *sdp.source,
			inviteHash: r.inviteHash,
		}

		if logMode > 0 {
			fmt.Println("callJoinPayload -> ", payload)
		}
		joinGroupCallResult := r.joinVoiceCall(payload)
		if joinGroupCallResult == nil || (*joinGroupCallResult).transport == nil{
			_ = peerConnection.Close()
			fmt.Println("No transport found")
			return false
		}
		sessionId := int(time.Now().Unix())
		conference := Conference{
			sessionId: sessionId,
			transport: *(*joinGroupCallResult).transport,
			ssrcs: []Ssrc{{ssrc: *sdp.source, isMain: true}},
		}
		err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP: SdpBuilder().fromConference(conference, true),
		})
		peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			go func() {
				ticker := time.NewTicker(time.Second * 3)
				for range ticker.C {
					errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
					if errSend != nil {
						fmt.Println(errSend)
					}
				}
			}()
			codec := track.Codec()
			if strings.EqualFold(codec.MimeType, webrtc.MimeTypeOpus) {
				fmt.Println("Got Opus track, saving to disk as output.opus (48 kHz, 2 channels)")
				saveToDisk(oggFile, track)
			} else {
				fmt.Println("UNKNOWN MIME TYPE")
			}
		})
		/*err = peerConnection.SetRemoteDescription(offer)
		if err != nil {
			_ = peerConnection.Close()
			fmt.Println(err)
			return false
		}
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			panic(err)
		}
		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			panic(err)
		}*/
		gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
		<-gatherComplete
		//fmt.Println(Encode(*peerConnection.LocalDescription()))
		//r.audioTrack = audioTrack
		select {}
		return true
	}else{
		return false
	}
}

func (r *RTCPeerConnectionClient) Track() *webrtc.TrackLocalStaticSample {
	return r.audioTrack
}
func saveToDisk(i media.Writer, track *webrtc.TrackRemote) {
	defer func() {
		if err := i.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			panic(err)
		}
		if err := i.WriteRTP(rtpPacket); err != nil {
			panic(err)
		}
	}
}

func (r *RTCPeerConnectionClient) joinVoiceCall(params JoinVoiceCallParams) *JoinVoiceCallResponse {
	var body = client.requestData(map[string]interface{}{
		"chat_id": params.chatId,
		"fingerprint" : params.fingerprint,
		"hash" : params.hash,
		"setup" : params.setup,
		"pwd" : params.pwd,
		"ufrag" : params.ufrag,
		"source" : params.source,
		"invite_hash" : params.inviteHash,
	}, "/request_join_call")
	var bodyResult map[string]interface{}
	err := json.Unmarshal([]byte(*body), &bodyResult)
	if err != nil {
		return nil
	}
	result := bodyResult["result"].(map[string]interface{})
	return JoinVoiceCallResponse{}.Parse(result)
}