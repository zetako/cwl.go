* runner.go

CWL 的运行环境，提供以下对象/接口

* NewEngine(...) Engine : 从 cwl 描述创建出执行器对象
* Engine 执行器对象
    * Run() 运行工作流 
        * MainProcess(...) Process : 返回抽象运行过程（根过程）
        * Steps() []Process : 返回子过程列表, 主要用于调试分析
        * Preparation() []Process :返回可以开始运行的子过程
        * CommitStep(stepid string,Values) 设置中间运行结果
* Process  运行过程对象
    * Command() : 解析出执行命令（只对CommandLineTool 过程有效）,供 Executor 使用 
    * ... ...
* Executor  CommandLineTool 运行器
    * Run(Process) : 执行过程
    * Outputs() Values :获取运行结果

    
-----

Engine.Run 按以下流程执行(示意)：

```go
rootProcess := Engine.MainProcess()
nextProcess := rootProcess.Preparation() 
while len(nextProcess) > 0 {
  for _, step := range nextProcess {
    out := Engine.RunStep(step)
    // 若 Step 为 commandLineTool , 则调用 Executor.Run
    Engine.CommitStep(step.id, out)
  }
  nextProcess = rootProcess.Preparation() 
}
return Engine.Outputs()
```

----- Advance 

可以通过以下接口修改默认行为：

* Engine 执行器对象
    * SetImporter(Importer) 设置非默认 Importer
    * SetExecutor(exec) 设置非默认执行器模块
    * SetFilesystem(fs) 设置非默认文件系统模块
* Process 
    * SetExecutor(Importer) 设置默认执行器（当前过程以及子过程）
* Importer 引用器，用于根据名称加载CWL文档,默认 Importer 从本地路径加载
* Filesystem 文件系统模块，默认采用当前文件系统

------- Dev Refs

https://common-workflow-lab.github.io/CWLDotNet/reference/CWLDotNet.File.html

 
