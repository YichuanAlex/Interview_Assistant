下面按“数据闭环周期”拆解。你的岗位更像 Seed Robotics 的 Data Infra / Data Engine：不是单纯写脚本，而是把机器人数据从采集、清洗、标注、质检、管理、训练、评估、回流串成可复用、可追踪、可规模化的系统。

一、端侧采集：把“机器人现场”变成可训练数据

端侧采集的对象不是单一视频，而是一个 episode。一个完整 episode 至少包含：

视觉：头部相机、腕部相机、环境相机、多视角视频帧。

动作：关节角、末端执行器 pose、夹爪开合、底盘速度、控制频率。

语言：任务指令，例如“把杯子放到架子上”。

环境元信息：机器人型号、相机参数、场景 ID、物体类别、光照、背景、操作者、采集时间。

结果信息：任务是否成功、失败原因、碰撞、卡顿、遮挡、是否中断。

所以平台首先要定义统一 schema。机器人数据不能只存 video_path，而要存：

episode_id
robot_type
task_instruction
camera_id
timestamp
fps
frame_path
action_chunk
joint_state
end_effector_pose
gripper_state
scene_metadata
quality_score
label_status
data_source
dataset_version

这一层的工程重点是“时序对齐”。视频是 30fps，控制可能是 10Hz、20Hz、50Hz，传感器还有延迟。平台要做 timestamp alignment，把每个视觉帧对应到最近的 robot state/action。否则训练 VLA 时会出现“看到的状态”和“执行的动作”错位。

二、原始数据入湖：先保证可查、可复现、可追踪

采集后不会直接进训练，而是进入数据湖或 HDFS/对象存储。文档里提到 DataSeek 是 Seed 的数据探索与管理平台，目标是解决数据分散、难发现、标准不一致的问题，并管理合版数据、应用回流、标注数据、训练数据、论文数据等资产。

机器人数据平台这里要做几件事：

第一，原始数据不可覆盖。raw episode 一旦入库，后续所有清洗、切片、重标注、过滤都应该生成 derived dataset，而不是修改原始数据。

第二，每个数据集要有元数据。包括来源、采集设备、采集任务、机器人平台、模态、标签类型、数据量、质量分布、是否可训练、是否合规。

第三，建立 dataset card。类似：

数据集名称：robotics_pick_place_v3
来源：真实机器人采集 / 仿真生成 / 人类第一视角视频
任务类型：pick-place / folding / navigation
模态：multi-view video + action + language
样本量：episodes / frames / hours
质量分布：1-5
训练用途：VLA SFT / world model / evaluation
版本：v1.0, v1.1
血缘：由哪些 raw datasets 派生

这和你材料里提到的“数据集详情预览和可视化、查看标签、相关人员、样例数据、能力、HDFS 地址”等能力是一致的。

三、自动预处理：把长视频变成结构化训练样本

机器人原始视频通常不能直接训练。需要一组数据处理算子。

第一类是视频基础处理：

抽帧：按 fps=1、2、5 或关键帧抽取。

切片：把长 episode 切成 subtask，例如“伸手 → 抓取 → 抬起 → 放置”。

去重：删除重复帧、重复 episode、近似任务。

质量检测：抖动、模糊、遮挡、黑屏、掉帧、时长异常。

格式统一：mp4、jpg、webp、json、parquet、hive 表统一。

第二类是机器人专用处理：

轨迹平滑检查：判断 action 是否有异常跳变。

碰撞检测：通过视觉或 robot log 判断是否撞击。

夹爪状态识别：open / close / holding。

物体接触状态：是否抓住，是否滑落。

坐标变换：camera frame、robot base frame、world frame 之间转换。

第三类是模型辅助预标注：

autoclip：自动判断动作步骤起止时间。

video caption：给每段视频生成动作描述。

异常检测：画面抖动、机械臂摇晃、长时间停顿、碰撞、物体卡住。

物体位置输出：输出 mask、bbox、关键点、机械臂重合度。

你截图里的“autoclip / 视频 caption / 视频异常检测 / 物体位置坐标输出”本质就是这一层。

工程上，这些算子应该做成 pipeline node，而不是散落的脚本。每个算子有标准输入输出，例如：

input: raw_episode
operator: video_caption_v2
params: fps=1, model=seed_2.0_lite
output: caption_segments.json

input: caption_segments.json + robot_log
operator: autoclip
output: subtask_segments.json

input: frames + robot_mask
operator: quality_checker
output: quality_report.json

文档也明确提到 DataSeek/相关数据平台需要支持去重、行列拆分、相似分析、过滤筛选、版本管理、质量分析和安全合规评估等能力。

四、标注与质检：从“可用数据”变成“高价值训练数据”

机器人数据最贵的部分是标注。平台会尽量做“模型预标 + 人工确认 + 质量抽检”。

典型流程：

模型先生成 caption、动作分段、bbox、质量标签。

人工只检查关键字段，而不是从零标注。

质检人员抽样复核，计算一致率。

对低一致率任务回退重标。

最终写入 labeled dataset。

对机器人任务，标注内容通常包括：

动作步骤名称。

每一步开始/结束时间。

任务是否成功。

失败原因。

关键物体位置。

机械臂是否遮挡。

是否碰撞。

是否掉落。

是否中断。

是否满足任务指令。

质量分 1-5。

这里最重要的是质量评分。你发的 PI/π0.7 内容里提到质量标签可以作为 prompt 输入，推理时设为高质量来引导模型输出更优动作。这个思路对 Seed Robotics 数据平台也很关键：失败数据不是垃圾，而是要保留失败类型、失败阶段、失败原因，用于后训练、对比学习或 hindsight relabeling。

五、数据资产管理：DataSeek 这一层解决“找数、用数、追数”

DataSeek 的角色可以理解为机器人数据的资产目录 + 检索系统 + 数据治理平台。它不是训练框架本身，而是让算法、数据、运营能找到、理解、加工、复用数据。

你给的文档里提到 DataSeek 支持关键词检索、向量检索、深度检索，也支持观察目标语料在训练数据 pipeline 中的留存情况。

对应到机器人数据，就是：

“找所有双臂机器人折衣服失败的数据。”

“找所有夹爪抓取透明杯失败的数据。”

“找所有腕部相机遮挡严重但任务成功的数据。”

“找 UR5e 在厨房场景的 pick-place 数据。”

“找质量分 >=4、无碰撞、动作完整的放置任务。”

技术上会用两类索引：

结构化索引：robot_type、task_type、scene、success、quality_score、duration、object_class。

语义索引：caption embedding、instruction embedding、video embedding、frame embedding。

所以平台开发可能涉及：

Elasticsearch / OpenSearch 做全文检索。

向量数据库或 ANN 索引做相似检索。

Hive / Spark / Flink 做大规模数据处理。

对象存储/HDFS 存 raw media。

Parquet/JSONL 存训练样本。

元数据 DB 存 dataset card、版本、血缘。

六、数据版本与血缘：回答“这次模型提升到底来自哪批数据”

这是数据平台最关键、也最容易被忽视的部分。

文档中提到样本库/数据平台需要记录血缘，并且要能观察某条数据是否通过每个处理阶段、在哪个环节被过滤、最终是否进入训练。

机器人场景中，血缘大概是：

raw_episode_v1
→ cleaned_episode_v1
→ clipped_segments_v2
→ captioned_segments_v2
→ quality_filtered_dataset_v3
→ train_mix_robotics_202606
→ VLA_model_ckpt_001
→ eval_report_001

每一步都要记录：

输入数据集 ID。

处理算子版本。

模型版本。

参数配置。

过滤规则。

输出数据集 ID。

执行时间。

负责人。

失败日志。

这样模型效果变化时才能定位：

是新数据有效？

是 caption 质量提升？

是过滤规则更严格？

是仿真数据占比过高？

是某类机器人数据污染了训练集？

文档也提到当前很多数据处理是 case-by-case，Pyspark map/filter 自定义逻辑很多，不容易复用。 所以平台工程的价值就是把这些逻辑抽象成可配置、可复用、可追踪的算子。

七、模拟数据生成：用仿真和生成模型补足真实数据缺口

机器人真实采集成本高，因此 Seed Robotics 会高度依赖模拟数据和生成式数据。

你截图里提到的方向包括：

轨迹数据仿真生视频。

环境泛化生成。

首图 + 文本生视频。

多视角仿真生成。

3D 物体仿真生成。

这些可以分成三类。

第一类：物理仿真数据。

用 Isaac Sim / Mujoco / Omniverse / 自研仿真环境生成 robot trajectory。输出包含 RGB、depth、segmentation mask、camera pose、object pose、action、reward、success flag。

优点是标签天然准确，bbox、mask、深度、pose 都可以自动得到。缺点是 sim-to-real gap。

第二类：生成模型增强。

比如保持机械臂动作不变，把厨房背景换成超市/冰箱/办公室；保持物体类别不变，替换颜色、纹理、光照；用图像/视频生成模型扩展少见场景。

这类数据要特别注意一致性：

机械臂不能变形。

相机视角不能漂移。

物体位置不能不合理。

动作轨迹不能违反物理规律。

多视角之间要一致。

第三类：3D 资产生成。

根据图像或文本生成 glb/obj/usd/usdz 等 3D 资产，用于仿真场景搭建。你截图中 3D 物体生成就是这个方向。

平台要支持：

3D asset registry。

材质/PBR 信息。

尺度归一化。

碰撞体生成。

可抓取点估计。

导入仿真引擎。

和真实物体类别映射。

八、训练数据构建：从数据资产变成训练样本

不同模型需要不同数据格式。

如果是 VLM/VLA 训练，样本可能长这样：

{
"instruction": "put the cup on the shelf",
"images": ["cam1_t.jpg", "cam2_t.jpg", "wrist_t.jpg"],
"state": [...],
"subgoal_image": "future_t+4s.jpg",
"action_chunk": [[...], [...]],
"robot_type": "UR5e",
"quality": 5,
"metadata": {...}
}

如果是 world model 训练，样本可能是：

current_frame + instruction → future_subgoal_frame

如果是 caption/autoclip 模型训练，样本是：

video segment → structured action steps with timestamps

如果是异常检测模型训练，样本是：

video segment → abnormal_type + abnormal_time + reason

文档里提到 Merlin 数据湖支持预览、debug，并且可以和 byted_streaming、veomni 打通，通过 yaml 做 multi-source 训练。 这说明平台最终要把数据整理成训练系统能直接消费的格式，而不是只停留在“数据展示”。

九、模型训练与评估：数据飞轮的后半段

训练不是终点。平台要记录：

这次训练用了哪些数据集。

每个数据集占比多少。

真实数据、仿真数据、人类视频、失败数据各占多少。

质量分分布如何。

哪些数据被过滤掉。

训练后在哪些 benchmark 提升。

哪些任务下降。

典型评估维度：

任务成功率。

动作平滑度。

碰撞率。

完成时间。

泛化到新物体。

泛化到新场景。

跨机器人迁移。

语言指令理解。

失败恢复能力。

如果某模型在“透明杯抓取”任务失败率高，平台应该能反向检索：

有没有透明杯数据？

质量如何？

是否只来自仿真？

是否缺腕部视角？

是否没有失败样本？

是否标注不一致？

这就是数据闭环。

十、平台管道开发：实习生最可能实际做什么

结合岗位“机器人数据开发实习生”，你实际可能做的不是发明 VLA，而是这些工程任务：

写数据处理算子：抽帧、切片、caption 后处理、bbox 格式转换、轨迹格式转换、quality report 生成。

写 pipeline DAG：把采集、清洗、标注、过滤、入库、同步训练串起来。

写数据 schema 和校验器：检查字段缺失、时间戳错位、action 维度不一致、视频损坏。

写检索和可视化：支持按任务、机器人、场景、质量分筛选数据。

写数据质量分析：统计成功率、失败原因、模态缺失率、重复率、场景分布、物体分布。

写训练导出工具：把 DataSeek 数据集导出成 jsonl/parquet/yaml，供 Merlin/veOmniverse/训练平台使用。

写血缘记录：记录每个 dataset 是由哪些 raw data 和哪些 operator 生成的。

写仿真数据转换：把 trajectory json、USD/OBJ/GLB、rendered video 转成统一训练格式。

一句话总结：

这个岗位的核心是“让机器人数据可规模生产、可检索、可评估、可追踪、可直接进入训练”。

你面试时可以这样表达：

“我理解 Seed Robotics 数据平台的工作不是单点数据清洗，而是建设具身模型的数据飞轮。端侧采集得到多模态 episode 后，平台通过统一 schema、自动化处理算子、模型预标注、人工质检、数据版本和血缘管理，把 raw data 转换成可训练、可评估、可复用的数据资产。之后再联动训练平台，把模型评估结果回流到数据发现和数据生产环节，持续补齐低覆盖、高价值、失败恢复和泛化场景的数据。”
