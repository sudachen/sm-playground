package japi

type echo struct {
	Msg struct {
		Value string `json:"value"`
	} `json:"msg"`
}

type echoRes struct {
	Msg struct {
		Value string `json:"value"`
	} `json:"msg"`
}

func (a *Agent) Echo(s string) (r string, err error) {
	i, o := echo{}, echoRes{}
	i.Msg.Value = s
	err = a.post("/node/echo",&i,&o)
	if err != nil {
		r = o.Msg.Value
	}
	return
}
