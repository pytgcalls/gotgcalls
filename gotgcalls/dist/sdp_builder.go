package main

import (
	"fmt"
	"strings"
)

func SdpBuilder() *SdpBuilderClient{
	return &SdpBuilderClient{
		lines: []string{},
		newLine: []string{},
	}
}

type SdpBuilderClient struct{
	lines []string
	newLine []string
}

func (r *SdpBuilderClient) fromConference(conference Conference) string {
	r.addConference(conference)
	return r.finalize()
}

func (r *SdpBuilderClient) addConference(conference Conference) {
	ssrcs := conference.ssrcs
	sessionId := conference.sessionId
	r.addHeader(sessionId, ssrcs)
	for ssrc := range ssrcs{
		r.addSsrcEntry(ssrcs[ssrc], conference.transport)
	}
}
func (r *SdpBuilderClient) addHeader(sessionId int, ssrcs []Ssrc) {
	r.add("v=0")
	r.add(fmt.Sprintf("o=- %d 2 IN IP4 0.0.0.0", sessionId))
	r.add("s=-")
	r.add("t=0 0")
	var genSsrc []string
	for ssrc := range ssrcs{
		genSsrc = append(genSsrc, r.toAudioSsrc(ssrcs[ssrc]))
	}
	audioSsrcs := strings.Join(genSsrc, " ")
	r.add(fmt.Sprintf("a=group:BUNDLE %s", audioSsrcs))
	r.add("a=ice-lite")
}

func (r *SdpBuilderClient) add(line string) {
	r.lines = append(r.lines, line)
}
func (r *SdpBuilderClient) push(line string) {
	r.newLine = append(r.newLine, line)
}
func (r *SdpBuilderClient) toAudioSsrc(ssrc Ssrc) string {
	if ssrc.isMain {
		return "0"
	}
	return fmt.Sprintf("audio%d", ssrc.ssrc)
}
func (r *SdpBuilderClient) addSsrcEntry(entry Ssrc, transport Transport) {
	ssrc := entry.ssrc
	var isMain int
	if entry.isMain {
		isMain = 1
	}else{
		isMain = 0
	}
	r.add(fmt.Sprintf("m=audio %d RTP/SAVPF 111 126", isMain))
	if entry.isMain{
		r.add("c=IN IP4 0.0.0.0")
	}
	r.add(fmt.Sprintf("a=mid:%s", r.toAudioSsrc(entry)))
	if entry.isRemoved != nil && *entry.isRemoved {
		r.add("a=inactive")
		return
	}
	if entry.isMain{
		r.addTransport(transport)
	}
	r.add("a=rtpmap:111 opus/48000/2")
	r.add("a=rtpmap:126 telephone-event/8000")
	r.add("a=fmtp:111 minptime=10; useinbandfec=1; usedtx=1")
	r.add("a=rtcp:1 IN IP4 0.0.0.0")
	r.add("a=rtcp-mux")
	r.add("a=rtcp-fb:111 transport-cc")
	r.add("a=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level")
	if entry.isMain{
		r.add("a=sendrecv")
	}else{
		r.add("a=sendonly")
		r.add("a=bundle-only")
	}
	r.add(fmt.Sprintf("a=ssrc-group:FID %d", ssrc))
	r.add(fmt.Sprintf("a=ssrc:%d cname:stream%d", ssrc, ssrc))
	r.add(fmt.Sprintf("a=ssrc:%d msid:stream%d audio%d", ssrc, ssrc, ssrc))
	r.add(fmt.Sprintf("a=ssrc:%d mslabel:audio%d", ssrc, ssrc))
	r.add(fmt.Sprintf("a=ssrc:%d label:audio%d", ssrc, ssrc))
}

func (r *SdpBuilderClient) addTransport(transport Transport) {
	r.add(fmt.Sprintf("a=ice-ufrag:%s", transport.ufrag))
	r.add(fmt.Sprintf("a=ice-pwd:%s", transport.pwd))
	fingerprints := transport.fingerprints
	for fingerprint := range fingerprints{
		r.add(fmt.Sprintf("a=fingerprint:%s %s", fingerprints[fingerprint].hash, fingerprints[fingerprint].fingerprint))
		r.add("a=setup:passive")
	}
	candidates := transport.candidates
	for candidate := range candidates{
		r.addCandidate(candidates[candidate])
	}
}
func (r *SdpBuilderClient) addCandidate(c Candidate) {
	r.push("a=candidate:")
	r.push(fmt.Sprintf("%s %s %s %s %s %s typ %s", c.foundation, c.component, c.protocol, c.priority, c.ip, c.port, c.typeConn))
	r.push(fmt.Sprintf(" generation %s", c.generation))
	r.addJoined()
}
func (r *SdpBuilderClient) addJoined() {
	r.add(strings.Join(r.newLine, ""))
	r.newLine = []string{}
}
func (r *SdpBuilderClient) finalize() string {
	return r.join() + "\n"
}
func (r *SdpBuilderClient) join() string{
	return strings.Join(r.lines, "\n")
}