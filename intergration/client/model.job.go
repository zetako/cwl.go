package client

import "encoding/json"

// Job 运行中的作业
type Job struct {
	ID            int        `json:"id"`             // 作业编号
	UserName      string     `json:"user_name"`      // 系统用户名
	GroupName     string     `json:"group_name"`     // 租户名称
	ClusterName   string     `json:"cluster_name"`   // 集群名称
	PartitionName *string    `json:"partition_name"` // 分区名称
	ClusterJobID  string     `json:"cluster_job_id"` // 集群作业编号
	Name          string     `json:"name"`           // 作业名称
	WorkDir       *string    `json:"work_dir"`       // 作业临时目录
	Status        JobStatus  `json:"status"`         // 作业运行状态，取值为0：排队中；1：运行失败；2运行中
	CreatedAt     Timestamp  `json:"created_at"`     // 提交时间
	UpdatedAt     Timestamp  `json:"updated_at"`     // 提交时间
	StartedAt     *Timestamp `json:"started_at"`     // 开始时间
	MaxRunTime    int        `json:"max_run_time"`   // 最长运行时间，单位小时，0则不限制
	EndAt         *Timestamp `json:"end_at"`         // 结束时间
	CpuPrice      float64    `json:"cpu_price"`      // CPU价格
	CpuUsed       int        `json:"cpu_used"`       // CPU使用量
	GpuPrice      float64    `json:"gpu_price"`      // GPU价格
	GpuUsed       int        `json:"gpu_used"`       // GPU使用量
	MemoryPrice   float64    `json:"memory_price"`   // 内存价格
	MemoryUsed    int64      `json:"memory_used"`    // 内存使用量
	Node          int        `json:"node"`           // 节点数
	Type          JobType    `json:"type"`           // 作业的类型，取值为0：HPC；1：K8S
	IP            string     `json:"ip"`             // 作业所在IP地址
	ExitCode      int        `json:"exit_code"`      // 退出码
	Reason        string     `json:"reason"`         // 退出说明
	SuspendedTime int64      `json:"suspended_time"` // 挂起时间
	// v3 need fields
	ParentJobId     string `json:"parent_job_id"`    // 父作业编号
	AppName         string `json:"app_name"`         // 应用名称`
	ConcurrentIndex int    `json:"concurrent_index"` //工作流使用
	NodeName        string `json:"node_name"`        //Pod的运行节点
	// dynamic extends
	UID             int             `json:"uid,omitempty"`             // 系统用户ID
	UUID            string          `json:"uuid,omitempty"`            // Task uuid
	JobFee          float64         `json:"job_fee"`                   //作业消耗机时
	NodePrice       float64         `json:"node_price,omitempty"`      //节点价格，用于HPC 作业
	NodesList       []string        `json:"nodes_list,omitempty"`      //节点【或pods】列表 ， 动态获取
	Proxies         []Proxy         `json:"proxies,omitempty"`         // 代理或网络入口信息，动态获取
	Extends         json.RawMessage `json:"extends,omitempty"`         // 扩展信息，动态获取
	Children        []Job           `json:"children,omitempty"`        // 子作业信息
	ChildrenRunning []Job           `json:"childrenRunning,omitempty"` // 子作业信息
	NodeTime        int64           `json:"node_time,omitempty"`       // 节点时间，用于 SLURM 作业的中间计费，已经考虑过节点数
	Steps           int             `json:"steps,omitempty"`           // resize 或 restart 的 次数 ， 1 /0 均表示没有重新分配
	ResizeAt        int64           `json:"resizeAt,omitempty"`        // resizeAt , 从 agent 获取

	DVersion       string `json:"dversion,omitempty"`       // debugVersion, 从 agent 获取
	IterationIndex int    `json:"iterationIndex,omitempty"` // 继承自对应的 Workflow.Task
}

type JobStatus int

const (
	// Failing    专门用于 提醒用户主动查看作业现场；关闭后成为 Failed
	JobStatusFailing        JobStatus = iota - 3
	JobStatusUnknown                  // -2
	JobStatusRunningFail              // -1
	JobStatusPending                  //0
	JobStatusSuspended                //1 挂起
	JobStatusRunning                  //2 // > 2 的状态都是结束状态，非结束状态不在使用 >2 的状态码
	JobStatusSuccess                  //3
	JobStatusFailed                   //4
	JobStatusFailedSubmit             //5
	JobStatusSuspendedUnuse           //6 挂起
	JobStatusCanCelled                // 7 取消
	JobStatusPartialSuccess           // 8 部分成功

)

type JobType int

// JobType 的层次
const (
	JobTypeUnknown                JobType = iota // 0
	JobTypeHpcBatchDirect                        // 1
	JobTypeK8sJobDirect                          // 2 用量从 pod 获取
	JobTypeK8sDeploymentDirect                   // 3
	JobTypeK8sPod                 JobType = 64   // 4
	JobTypeHpcStep                JobType = 65   // 4
	JobTypeHpcCollapsed           JobType = 66   //
	JobTypeK8sJobCollapsed        JobType = 67   // 用量直接 获取
	JobTypeK8sDeploymentCollapsed JobType = 68
	JobTypeHpcBatch               JobType = 128
	JobTypeK8sJob                 JobType = 129
	JobTypeK8sDeployment          JobType = 130
	JobTypeWorkflow               JobType = 192
	// 其他情况 > 255
)
