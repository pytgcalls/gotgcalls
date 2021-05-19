package main

type Sdp struct {
	fingerprint *string
	hash *string
	setup *string
	pwd *string
	ufrag *string
	source *int
}
type JoinVoiceCallParams struct {
	chatId int
	fingerprint string
	hash string
	setup string
	pwd string
	ufrag string
	source int
	inviteHash *string
}
type Conference struct {
	sessionId int
	transport Transport
	ssrcs []Ssrc
}
type Ssrc struct {
	isMain bool
	isRemoved *bool
	ssrc int
}
type JoinVoiceCallResponse struct {
	transport *Transport
}
func (j JoinVoiceCallResponse) Parse(result map[string]interface{}) *JoinVoiceCallResponse{
	transport := result["transport"]
	if transport != nil{
		var fingerprints []Fingerprint
		transportConverted := transport.(map[string]interface{})
		fingerprintReturn := transportConverted["fingerprints"].([]interface{})
		for fingerprint := range fingerprintReturn{
			fingerprints = append(fingerprints, Fingerprint{
				hash: fingerprintReturn[fingerprint].(map[string]interface{})["hash"].(string),
				fingerprint: fingerprintReturn[fingerprint].(map[string]interface{})["fingerprint"].(string),
			})
		}
		var candidates []Candidate
		candidateReturn := transportConverted["candidates"].([]interface{})
		for candidate := range candidateReturn{
			candidates = append(candidates, Candidate{
				generation: candidateReturn[candidate].(map[string]interface{})["generation"].(string),
				component: candidateReturn[candidate].(map[string]interface{})["component"].(string),
				protocol: candidateReturn[candidate].(map[string]interface{})["protocol"].(string),
				port: candidateReturn[candidate].(map[string]interface{})["port"].(string),
				ip: candidateReturn[candidate].(map[string]interface{})["ip"].(string),
				foundation: candidateReturn[candidate].(map[string]interface{})["foundation"].(string),
				id: candidateReturn[candidate].(map[string]interface{})["id"].(string),
				priority: candidateReturn[candidate].(map[string]interface{})["priority"].(string),
				typeConn: candidateReturn[candidate].(map[string]interface{})["type"].(string),
				network: candidateReturn[candidate].(map[string]interface{})["network"].(string),
			})
		}
		return &JoinVoiceCallResponse{
			transport: &Transport{
				ufrag: result["transport"].(map[string]interface{})["ufrag"].(string),
				pwd: result["transport"].(map[string]interface{})["pwd"].(string),
				fingerprints: fingerprints,
				candidates: candidates,
			},
		}
	}else{
		return &JoinVoiceCallResponse{
			transport: nil,
		}
	}
}
type Transport struct {
	ufrag string
	pwd string
	fingerprints []Fingerprint
	candidates []Candidate
}
type Fingerprint struct {
	hash string
	fingerprint string
}
type Candidate struct {
	generation string
	component string
	protocol string
	port string
	ip string
	foundation string
	id string
	priority string
	typeConn string
	network string
}