package main

import (
	"encoding/json"
	"fmt"
	"github.com/pion/webrtc/v3"
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
	ctxIceConnected := make(chan bool)
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err == nil{
		r.rtcConnection = peerConnection
		peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
			if logMode > 0 {
				fmt.Printf("IceConnection State has changed to %s \n", connectionState.String())
			}
			if connectionState == webrtc.ICEConnectionStateConnected {
				ctxIceConnected <- true
			}
		})
		audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion")
		if err != nil {
			return false
		}
		rtpSender, err := peerConnection.AddTrack(audioTrack)
		if err != nil {
			_ = peerConnection.Close()
			return false
		}
		go func() {
			rtcpBuf := make([]byte, 1500)
			for {
				if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
					return
				}
			}
		}()
		peerConnection.OnTrack(func(remote *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			fmt.Println(remote)
		})
		offer, err := peerConnection.CreateOffer(nil)
		if err != nil {
			_ = peerConnection.Close()
			return false
		}
		err = peerConnection.SetLocalDescription(offer)
		if err != nil {
			_ = peerConnection.Close()
			return false
		}
		sdp := r.parseSdp(offer.SDP)
		if sdp.ufrag == nil || sdp.pwd == nil || sdp.hash == nil || sdp.fingerprint == nil || sdp.source == nil {
			return false
		}
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
			SDP: SdpBuilder().fromConference(conference),
		})
		if err != nil {
			_ = peerConnection.Close()
			return false
		}
		r.audioTrack = audioTrack
		return true
	}else{
		return false
	}
}

func (r *RTCPeerConnectionClient) Track() *webrtc.TrackLocalStaticSample {
	return r.audioTrack
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