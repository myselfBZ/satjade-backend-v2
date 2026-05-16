package events



var _ ClientSentEvent = (*AuthEvent)(nil)


type AuthEvent struct {
	Token string `json:"token"`
}

func (a *AuthEvent) isClientSentEvent() {}
