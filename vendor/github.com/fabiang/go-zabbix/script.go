package zabbix

const (
	ScriptTypeScript  = 0
	ScriptTypeIPMI    = 1
	ScriptTypeSSH     = 2
	ScriptTypeTelnet  = 3
	ScriptTypeWebhook = 5

	ScriptScopeActionOperation     = 1
	ScriptScopeManualHostOperation = 2
	ScriptScopeManualEventAction   = 4

	ScriptExecuteOnAgent  = 0
	ScriptExecuteOnServer = 1
	ScriptExecuteOnProxy  = 2

	ScriptResponseSuccess = "success"
	ScriptResponseFailed  = "failed"
)

type Script struct {
	ScriptID    string `json:"scriptid"`
	Name        string `json:"name"`
	Type        int    `json:"type"`
	Command     string `json:"command"`
	Scope       int    `json:"scope"`
	ExecuteOn   int    `json:"execute_on"`
	Description string `json:"description"`
}

type ScriptExecuteRequest struct {
	ScriptID string `json:"scriptid"`
	HostID   string `json:"hostid"`
}

type ScriptExecutionResponse struct {
	Response string      `json:"response"`
	Value    string      `json:"value"`
	Debug    interface{} `json:"debug"`
}

func (c *Session) ScriptExecute(req ScriptExecuteRequest) (ScriptExecutionResponse, error) {
	response := ScriptExecutionResponse{}
	err := c.Get("script.execute", req, &response)
	return response, err
}
