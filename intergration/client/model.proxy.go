package client

// Proxy 代理信息
type Proxy struct {
	ID int `json:"id"` // 代理编号
	// 需要注意，前端的 name 长度限制需小于 32，数据库的长度限制需 = 前端限制 + job.name 的长度限制
	Name            string `json:"name"`                       // 代理名称
	UserName        string `json:"user_name"`                  // 系统用户名
	EntryName       string `json:"entry_name"`                 // 入口名称
	JobID           int    `json:"job_id"`                     // 作业编号
	Type            int    `json:"type"`                       // 代理方式，取值为1：固定端口；2：随机端口
	Protocol        int    `json:"protocol"`                   // 代理协议，取值为1：http；2：https；3：SSH；4：TCP
	Domain          string `json:"domain"`                     // 外部域名
	CertFile        []byte `json:"cert_file"`                  // HTTPS代理的ca证书
	KeyFile         []byte `json:"key_file"`                   // HTTPS代理ca证书的密钥
	SourcePort      int    `json:"source_port"`                // 源端口
	TargetPort      int    `json:"target_port"`                // 目的端口
	ServicePort     int    `json:"service_port,omitempty"`     // 内部 svc 使用的 port , 默认（缺省）和 TargetPort 一致； 用于扩展 proxy 数据结构的应用
	ServiceProtocol string `json:"service_protocol,omitempty"` // 内部 svc 使用的 protocol , 默认（缺省） 为 TCP ； 用于扩展 proxy 数据结构的应用
	IP              string `json:"ip"`                         // 目的IP
	Control         int    `json:"control"`                    // 验证级别，取值为1：不验证；2：登录用户；3：私有
	ShareCode       string `json:"share_code"`                 // 共享码
}

const (
	ProxyProtocolHTTP  int = 1
	ProxyProtocolHTTPS int = 2
	ProxyProtocolSSH   int = 3
	ProxyProtocolTCP   int = 4
)
