TL;DR12 月 25 日共识

左侧（数据收集+加工清洗+二次处理）数仓这边维护，保障数据的全面和质量 ，数据处理算子方面研发可以共同开发@Xiaolong Zhou

右侧样本库维护所有数仓+算法这边关于治理侧所有的数据集的元信息管理，涵盖治理范围的所有数据，运营和算法可以在平台进行数据的发现、圈选、挖掘，对数据进行管理并利用一些已有的算子做加工。  @Fenix Zhu

数据提需和验收流程以及标准，需要和业务对齐 @Xiaolong Zhou  @Fenix Zhu

现状和未来https://bytedance.larkoffice.com/sync/VZzNdRwqksWFqhbCfayc27HenkJ原型设计数据链路：样本库讨论版本This content is only supported in a Feishu Docs数据链路：样本库讨论版本总体思路建设目标1+3+N

1 套数据资产体系；治理相关的所有数据，都能在样本库检索发现（数据集检索+数据挖掘）衡量指标：模型（模型+rag）覆盖率，风险标签覆盖度，标出率，模型效果

3 套数据生产自动/半自动 pipeline；治理的数据生产线上化，可追踪数据血缘，给业务决策做辅助

N 个数据处理工具；数据加工和处理工具箱，运营可利用来生产数据，提升生产效率和质量衡量指标：运营自助率提升，生产效率提升建设节奏This content is only supported in a Feishu DocsQ1 OKR

Action1：数据资产：数仓样本基建完善度提升至90%+，模型迭代链路下的视频+直播覆盖度 100%。@Fenix Zhu

Task1：持续建设样本基建表，完善度提升至90%+  w/ @Xiaolong Zhou

Task2：例行送标 & 取数链路收敛链路，移除中间层，实现 100%从数仓取数 w/@Tao Liu

Task3：【探索向】建设样本实时流接入，支撑行业数据生产链路 w/@guanglian wang

Action2：数据管理：样本检索能力增强，新增相似检索和挖掘能力，标出率提升 xx%，运营自助率 xx%

Task1：样本可视化（卡片/详情展示，支持图集、视频、音频播放）：视频/直播内容三级风险标签覆盖率100% @Lingchen Li （视频）@Qianyu Pan （直播）@Licheng Xu （生态/行业）@Qinuo Li

Task2：样本检索能力：完善样本检索能力，数据召回率 > xx % @Fenix Zhu  @Steed

Task3：样本扩召能力：完善样本挖掘能力（相似检索、eva fewshot 扩召、query zeroshot 检索、高温采样、聚类去重等能力），支持运营调整干预参数，数据召回率 > xx%，标出率提升 xx%  @Fenix Zhu  @Steed  w/算法

Action3：数据生产链路：建设数据加工生产 pipeline能力，完善数据二次加工的算子工具箱，数据加工线上化率 x %

Task1：数据生产 pipeline 完成MVP 建设：跑通模型自动迭代链路数据生产的完整线上化（视频危险行为，直播不良导向）

Task2：schema标准化工具：每个业务方能自助完成样本schema 标准化加工

Task3：标准变更感知工具：感知风险标准（包括审核要素）变更，对接标准中心的变更通知，基于审核要素的增删改事件，标记样本是否需要重新标注或者废弃，在样本市场、样本集展现标准变更标签，并能筛选出标准变更的样本。

Action4：数据血缘：血缘中心对接，建立hive表 <--> 样本集（包含标注任务） <--> 模型版本（含训练任务）的血缘关系

Task1：建设样本集和数据源（比如hive表，MQ队列）的血缘关系，并上报给血缘中心

Task2：建设样本集和模型版本的血缘关系，并上报给血缘中心

Task3：页面里提供血缘入口，点击可查看样本集的血缘关系

This content is only supported in a Feishu Docs

背景在安全治理场景，大模型落地面临如下两个问题：

效果的的关键，是优质数据，越来越多的研究和实践表明，数据量不在多，而在精。针对治理场景的审核数据特征，怎么构建数据、从哪些维度构建，需要有一套数据工程体系来支撑。

数据横跨生产、基座训练、PE+RAG、SFT+RL 训练、评测上线等各个环节，如何协同算法、数仓、产品、运营高效的协作。

针对上述的问题，当前围绕下面的几个角度进行盘点

在各个环节流程下，业务生产的 pipeline 和SOP、数据资产和数据处理的原子能力等。

围绕数据的各个流程环节，盘点优质数据生产的策略、数据质量提升的手段等。总览This content is only supported in a Feishu Docs业务视角

数据生产 pipeline 建设（自动化、半自动化）

提升数据的质量和生产效率能力视角

一套数据资产管理体系（基建、存储和管理）

数据处理的工具箱：面向数据的生命周期的各个阶段，提供数据处理的算子能力业务视角数据生产pipelineThis content is only supported in a Feishu Docs面向业务主要有三种生产的 pipeline：

面向大模型风险治理场景：通过感知模型劣化、挖掘并做数据标注的形式积累数据，进而触发模型训练。

面向算法实验：一些新风险、新实验尝试，算法会在优质样本的基础上针对领域的差异性做采样、特征补齐或者其他标注/清洗逻辑。

面向基座模型迭代：需要大规模的产出业务 PA 数据、PAR 数据和一些基础数据如 Caption、OCR 等 VQA 数据。关键阶段This content is only supported in a Feishu Docs

关键阶段阶段说明角色输入输出评估标准存在问题预计解法数据收集

收集线上数据、合成数、标注数据

数仓

线上数据

数仓基建数据如：当日投稿、人审日志、机审日志、处置日志、标注、离线特征等暂无明确标准

覆盖度

数据源是否全覆盖

数据没有实时接入



加工清洗

分业务进行采样、特征的补齐、打标，方便后续筛选和二次加工

数仓

数仓基建数据

优质样本表

视频：视频、生态、AIGC

直播：图片流、音频流、图音综合流

评论：xxx 等暂无

量级

质量提升

实时性不足，当前只有天级优质样本表

字段和 tag 要齐全，现在不方便筛选取数 知识库数据源分类方案

数据本身的增强，1min 切片，3min 切片等

样本质量无明确的评估标准

数据处理算子工具比较多，各业务都有自己独特的方式，主要靠线下

二次加工依赖优质样本中的数据，进行二次加工，如生产 cot 或 格式转换

研发模型运营/算法数仓

优质样本表

数据应用的集合

暂无明确评估标准标注有 CoT 质量分（依赖于人工）机器生产的数据，暂无

通用问题

标注有 CoT 质量分（依赖于人工）机器生产的数据，暂无样本质量评价标准，缺乏预测仿真加工之后的数据的效果，每次都得进行实际训练之后或依赖算法自己的经验看效果，造成训练浪费

数据加工线下处理，无线上加工的 pipeline，无法跟踪

当前数据只是追加，无汰换机制面向自动迭代

挖掘效果评估问题

标注的人效看不清

数据源的类型混杂在一个标注任务里，比如来自于处置、申诉、模型分数较低的case等， 看不清哪批数据的有效性好或者差  ；

当标注有效性低于某个阈值时， 仅通知了标注模型训练poc，缺少解决和跟进的SLA；面向算法探索

自主取数训练之后，就上线模型，导致血缘无法追踪

特征不好加工面向基模

审核要素不方便获取，当前是 L5 的数据

数据应用





数据资产没有统一管理界面，数据发现效率低，数据维护和治理不方便

血缘无法追踪

看不清数据配比和使用方式，是否重复送标、正负例配比的样子











高质量数据生产实践

实现方案业务收益难样本挖掘

短视频基于 eva 的 few shot难样本挖掘送标快速从eva训练一个few-shot回扫模型，挖掘高价值的待标注数据送标[Image][Image][Image]

直播样本挖掘链路样本挖掘流程建设方案This content is only supported in a Feishu Docs

短视频基于eva模型回扫扩召链路，在未成年场景下acc和auc都提升2个点，rec@p95提升16个点（离线实验，线上效果待观测）[Image]线上6.28数据离线指标[Image]新训模型7.20数据离线指标

[Image][Image]

CoT 质检

智能标注CoT质检Agent[Image][Image]

[Image]

对线上模型的效果正相关采用质量分筛选高质量CoT数据，相比随机采样，训练模型得到效果更优，上线的模型的AUC和ACC都有不同程度提升（不良行为惩罚auc94->96，地方机关负面的acc 60->65%

有助于生产更高质量的CoT在无资质营销上，通过CoT质量分调整机标prompt，生产更高质量CoT，auc: 63.61%->67.20%。

视频 DQ Agent

AwemeVL-RL短视频强化学习基座与持续学习

主要思路

Workflow阶段：针对客观难度，定义特征维度和Policy维度的两个Judge Model（Seed-1.6）1. Feature-Level Judge Model：告知模型ground truth违规标签，模型根据视频内容判断判定为ground truth违规标签的难度；2. Policy-Level Judge Model：判断多个选项细则间的困惑程度；

Agent阶段【WIP】：增加工具数据量，通过Agentic RL训练范式，优化目标为RL基座模型reward收敛速度和reward指标[Image]

利用静态课程学习判断三种难度划分方法（logits、rollout、DQ Agent）的效果，DQ Agent方法对模型效果提升最明显，总体recall@p85/90/95较基础DAPO方法 +2.5～3.1pp

三种难度划分方法表现：DQ Agent = Rollout > Logits

加权平均Rollout和DQ Agent的分数对结果影响不大

DQ Agent用于去噪，对于剩余数据采用Rollout打分难度，该方法指标最优，但是提升幅度不大

Less is more：初步验证，DQ Agent方法可在训练数据量减少10%的基础上，达到优于logits方法的模型效果，减少数据训练时间（DQ Agent去噪声 + Rollout难度 + DAPO + 静态课程学习实验组）

人机协作的高质量 CoT 数据生产

在同一份数据集上支持多轮预标-人标：人去标注一些关键信息，LLM再用语言串起来。即人标级联选择题（审核要素 + 违规/豁免点），机器生产 CoT

This content is only supported in a Feishu DocsThis content is only supported in a Feishu DocsThis content is only supported in a Feishu Docs结合平台功能与离线分析，实现两次标注撞 diff + 头部同学精标，提升质量

已在「不良行为/惩罚」中应用，相比旧测试集，新测试集模型指标提升 58.65% → 63.08%

[Image]

能力视角数据资产基础表（底层数仓基建）

直播算法核心数据资产白皮书&FAQ 持续更新视频算法核心数据资产白皮书&FAQ 持续更新优质样本表（统一的训练、评估、基模的样本基建表）业务

统一样本 hive 表直播

图片流tnscdm_algo.dm_live_sample_image_high_quality_refined_di

音频流tnscdm_algo.ads_live_audio_review_high_quality_train_sample_di

图音综合流tnscdm_algo.ads_live_general_review_high_quality_train_sample_di

泛娱乐tnscdm_algo.dm_live_sample_entertainment_general_review_high_quality_di视频视频tnscdm_algo.dm_aweme_sample_video_high_quality_di

生态待建设

泛娱乐待建设

AIGC待建设评论评论tnscdm_algo.dm_aweme_sample_comment_high_quality_di标准维表视频+直播community_sensitive.dim_standard_clause_audit_element_detail_df

评论tnscdm_algo.dim_aweme_comment_policy_cleaned_book_df数据存储和管理数据湖基模训练存储采用的是 Merlin 的数据湖

支持预览，方便 debug，和 byted_streaming、veomni 打通，yaml 填写即可multi_source训练

Merlin 数据湖使用文档 审核基模-数据生产指南、Universal Prompt[Image][Image][Image]正式数据

{catalog}.{体裁}.{数据类型}{业务过程}{自定义tag}{训练or 验证}{版本}_{时间范围}

{体裁}：video, live, comment, toutiao, general

其他槽位的枚举值，看下面的说明临时数据{catalog}.{用户名}.{无所谓随便写}业务数据湖地址直播https://ml.bytedance.net/data/magnus/data_card?catalog=content_moderation_omni_ssd&region=cn视频https://ml.bytedance.net/data/magnus/data_card?catalog=content_moderation_omni_ssd&region=cn评论https://ml.bytedance.net/data/magnus/data_card?catalog=content_moderation_omni_ssd&region=cn样本库https://safe.bytedance.net/dev_portal/live/sample/sample-market[Image]样本库主要作用有两个

统一的数据服务层：对 eva 模型训练、sft 模型迭代、知识库案例库等偏产品形态上，提供了统一的数据服务，屏蔽掉底层的数据和存储差异；同时能记录数据使用血缘。

样本管理：可以按照三级标签和风险域的形式进行组织，可视化展示和管理样本。自定义的 hdfs 和 hive当前大部分算法同学自己仍然会使用 hive 或者 hdfs，未来这一部分会逐步减少。一次性的任务无所谓。数据处理工具箱当前的数据处理能力都是 case by  case 的去支持，而且使用成本很高，没有办法做到很便捷的复用。用户原声：from @Xiaolong Zhou「这里面主要是做一些基础数据处理，没啥算子复用性」「我们之前和架构搞过一套算子框架Deep，但目前线上定制化比较多，有很多代码是直接在任务里开发的」

from @Fubang Zhao「deep当时用的最多的，也是自定的map、filter，pyspark的处理逻辑，还是casebycase吧，不太好复用」「要注意“过度开发”的问题，有些清洗手段，其实还没有solid的结论，这块平台化的需求没那么高，还是再等等solid的结论会好一点」

from @De Cai「当前的数据处理也是定制化，逐步在做抽象，比如DQ Agent这种」

from @Tiangang Zhang「当前的数据处理都是定制化，数仓有在 aicore 搭建一些处理的工作流，如样本头图清洗，风险帧等」

from @Weizhao Li「数据处理都是算法自己写代码处理」

从数据处理能力来看，主要会这样划分：工具名维护方使用方式线上特征补齐

研发

通过数据模板 + puzzle 的形式，按题材粒度进行补齐当前实现了8 个业务身份下的数据特征补齐资源转存

研发/数仓

研发数仓都有，而且资源转存能力有三个地方

样本库有一层转存

数仓自己的 tos 转存

送到标注平台后，标注平台有一个转存视频抽帧xxx解法文本相似召回研发/算法根据自然语言召回相似的内容图片相似召回研发/算法根据图片或 ObjectID 召回相似内容根据直播间 id + 时间，召回 ObjectID研发

输入直播间 id和时间段，召回那个时间段的 ObjectID

直播图片流-优质样本正例选帧

数仓

https://douyin-ai.bytedance.net/detail?_withType=true&space=%22douyin_databp%22&id=%22app_7560952929138838318%22&from=%22app_7560952929138838318_container%22

利用大模型对图片流的帧进行风险帧选帧直播图音流-优质样本正例选帧

数仓

https://douyin-ai.bytedance.net/detail?_withType=true&space=%22douyin_databp%22&id=%22app_7579948195959737094%22&from=%22app_7560952929138838318_container%22

输入直播正例样本，进行样本数据清洗抖音-视频-国家安全-机标是否抖音-视频-偏激社会情绪和涉-机标是否抖音-视频-党和国家形象负面-机标是否

模型运营

https://douyin-ai.bytedance.net/detail?_withType=true&space=%22douyin_databp%22&id=%22app_7579876429351439155%22&from=%22app_7560952929138838318_container%22

https://douyin-ai.bytedance.net/detail?_withType=true&space=%22douyin_databp%22&id=%22app_7551986761077918464%22&from=%22app_7560952929138838318_container%22

https://douyin-ai.bytedance.net/detail?_withType=true&space=%22douyin_databp%22&id=%22app_7555426223757724442%22&from=%22app_7560952929138838318_container%22

对视频做数据清洗，筛选出来涉政相关的数据直播-例行机标 cot

数仓https://douyin-ai.bytedance.net/detail?_withType=true&space=%22douyin_databp%22&id=%22app_7576205497020205862%22&from=%22app_7560952929138838318_container%22

根据 top5 的标签，做 mcq，生成 cot 数据直播例行机标Caption

数仓

https://douyin-ai.bytedance.net/detail?_withType=true&space=%22douyin_databp%22&id=%22app_7573984668011023150%22&from=%22app_7560952929138838318_container%22

调用大模型，生成 Caption获取审核要素

数仓https://douyin-ai.bytedance.net/detail?_withType=true&space=%22douyin_databp%22&id=%22app_7566597294416366399%22&from=%22app_7560952929138838318_container%22

从风险知识库获取审核要素，并做格式转换视频选帧

数仓利用大模型，从抽帧里面挑选违规帧

https://douyin-ai.bytedance.net/detail?_withType=true&space=%22douyin_databp%22&id=%22app_7569534485767719690%22&from=%22app_7560952929138838318_container%22带画框的 grounding cot 生产

数仓+模型运营https://douyin-ai.bytedance.net/detail?_withType=true&space=%22douyin_databp%22&id=%22app_7566595661731269385%22&from=%22app_7560952929138838318_container%22[Image]

涉政领域的机标 CoT 数据合成

数仓+模型运营[Image]



重点建设事项

现状什么样？按之前的方式dfd串清楚，说清楚每个f是怎么干的？输入d经过f后有哪些变化？谁在干？数据漏斗标准怎么定义的？谁在定义？当前存在什么问题？

未来什么样？一起畅想下，从性能、成本、效率几个视角看下。

aicore 迁移的事情，现在机标和数据处理的算子。

基于上面的流程和问题，结合业务当前的痛点，定好优先级，挑重点问题解决，定里程碑。关键阶段阶段说明角色输入输出评估标准存在问题预计解法数据收集

收集线上数据、合成数、标注数据

数仓

线上数据

数仓基建数据如：当日投稿、人审日志、机审日志、处置日志、标注、离线特征等暂无明确标准

覆盖度

P0：数据源是否全覆盖，应该建设什么数据，包含什么数据，当前没有收口人拍板，无法定标准，要从数据消费视角让业务和算法做 提需和 验收check。

P2：数据没有实时接入

业务数据发现，需要业务提要接入数据源

加工清洗

分业务进行采样、特征的补齐、打标，方便后续筛选和二次加工

数仓

数仓基建数据

优质样本表

视频：视频、生态、AIGC

直播：图片流、音频流、图音综合流

评论：xxx 等

暂无

量级

质量提升

P0痛点 1：需求流程：哪些应该在数仓清洗，哪些应该在机审链路处理，主要是业务要能把需求提出来

P0 痛点 2：验收流程：当前清洗完数据之后，是由数仓来验收标注，理论上应该是让业务来验收样本是否满足诉求。

P1：数据处理算子工具比较多，各业务都有自己独特的方式，主要靠线下，搞一些工具需要研发帮忙做的；一次性的工具，需要 by case 去看哪些需要支持，看哪些需要工程开发。

P2：字段和 tag 要齐全，现在不方便筛选取数 知识库数据源分类方案

P2：实时性不足，当前只有天级优质样本表

样本数据需求流程和验收标准，定一下验收指标（量级和质量）

二次加工依赖优质样本中的数据，进行二次加工，如生产 cot 或 格式转换研发模型运营/算法数仓

优质样本表

数据应用的集合

暂无明确评估标准标注有 CoT 质量分（依赖于人工）机器生产的数据，暂无

通用问题

标注有 CoT 质量分（依赖于人工）机器生产的数据，暂无样本质量评价标准，每次都得进行实际训练之后或依赖算法自己的经验看效果

数据加工线下处理，无线上加工的 pipeline，无法跟踪血缘和漏斗

当前数据只是追加，无汰换机制

面向自动迭代

现在挖掘的清洗和入向量库库是工程在做，维护的复杂度也较高，后续这部分可以和数仓对齐

挖掘效果评估问题，没有明确的评估标准

标注痛点

标注的人效看不清

数据源的类型混杂在一个标注任务里，比如来自于处置、申诉、模型分数较低的case等， 看不清哪批数据的有效性好或者差  ；

当标注有效性低于某个阈值时， 仅通知了标注模型训练poc，缺少解决和跟进的SLA；

数据可视化现在Merlin是有的，但是必须得按照格式来，才能展示出来，需要有一个地方能够把历史的 hdfs 导入，预览，然后发起标注质检

面向算法探索

自主取数训练之后，就上线模型，导致血缘无法追踪

特征不好加工

定制化程度很高，算法的需求灵活，迭代快，容易出现本周生产的数据，下周就不能用

面向基模

审核要素不方便获取，当前是 L5 的数据，是否要保存到数据湖？

需求迭代快，灵活

从模型自动迭代链路，将数据生产 pipeline 完整线上化，追踪数据血缘和数据漏斗

依赖业务定汰换标准，才能做数据回溯和汰换数据应用





数据资产没有统一管理界面，数据发现效率低，数据维护和治理不方便

血缘无法追踪，看不清数据配比和使用方式，是否重复送标、正负例配比等



1+3+N

1 套数据资产体系；

3 套数据生产自动/半自动 pipeline；

N 个数据处理工具；

方向一：从数据资产上，持续建设，补齐优质样本表

方向二：面向不同场景的数据生产 pipeline 的完整线上化、自动化、例行化[Image]

方向三：数据管理方面，做好数据检索、可视化和信息维护，提升数据发现和使用效率

方向四：建设数据血缘（这部分从沟通来看算法和数仓没有感知血缘的价值，但是对于运营很有价值）

from seed 那里的「pipeline观察：支持观察目标语料在训练数据pipeline中的留存情况」观察内容：

某条语料是否通过了每个处理阶段

在哪个环节被过滤掉（如果被丢弃）

留存率统计（有多少数据最终进入训练）实际意义： 帮助理解为什么某些数据没有被用于训练

方向五：数据处理工具箱能力，沉淀一些通用的算子

参考历史整理：数据工程讨论相关调研学术论文相关【从零到灵读论文】大语言模型的数据工程https://github.com/yuleiqin/fantastic-data-engineering🌊Unleashing the Power of Data Tsunami: A Comprehensive Survey on Data Assessment and Selection for Instruction Tuning of Language Models

数据清洗

数据合成

数据筛选

数据增强论文里面主要讲的是数据处理和清洗的工具和算法，在做数据工具能力的时候，这些算法可以帮助打开思路。seeddataseek 平台DataSeek平台是专为Seed团队打造的智能数据平台，平台聚焦于智能数据探索与发现，帮助算法团队解决数据分散、难以发现、标准不一致等问题，快速定位并理解高价值数据，为模型训练提效。数据资产体系

合版数据、应用回流、标注数据、训练数据、论文数据

[Image]数据检索与发现

首页定期推荐高质量数据集

支持基于名称、描述等元数据进行全文搜索，提供灵活、多维度的标签和分类，便于业务快速定位[Image][Image]

数据集详情预览和可视化

支持查看标签、相关人员、样例数据、能力等丰富元信息，支持点击HDFS地址一键复制进行查询使用，支持关注数据集并收藏，可一键分享团队成员

[Image][Image]

[Image][Image]语料核查

语料检索：支持通过关键词检索、向量检索方式，查找并查看预训练语料详情、docid、原文本等信息

pipeline观察：支持观察目标语料在预训练数据pipeline中的留存情况 ，这个值得参考。

深度检索：支持通过自然语言实现预训练语料的深度检索[Image]能力

联动merlin，支持数据一键同步用于模型训练；

支持数据处理算子，如去重、行列拆分、相似分析、过滤筛选等；

支持数据版本管理与版本对比；

支持数据质量分析、安全合规评估等；商安

业务视角

业务数据生产 pipeline 的构建  --- 这里对应到我们未来大模型全机审运营干预的角度下，数据生产全自动 pipeline 建设

提升数据丰富度和高质量  ---- 这里对应到 cot 质量分、质检等能力建设

能力视角

数据统计分析、质量检查、清洗、安全、合成

各种丰富的算子能力[Image][Image]

modelcycle主要在做数据集管理[Image][Image]分析师和数仓协助构建了一些当时的领域数据集，然后这个数据工程产品就是在做数据集的编辑可视化、配比，然后能够衍生出来新的数据集。 重点在做模型的评估。但是这个后面就没有再做了。业界

基本流程[Image]每一步都有一个产品模块，或者业界有相关的产品

数据收集

爬虫

采买

可复用的公有的数据集

数据合成（cretel.ai、mostly ai、syntho 等公司）

文本数据合成

表格数据合成

数据转换

数据评估

数据标注（scale ai、整数智能等公司）

人工标注

半自动标注

自动标注

数据清洗/预处理（data-jucier等产品工具）

去重

规则过滤或启发式过滤

模型过滤

数据增强（图形增强（旋转、裁切、翻转、变换）、文本增强（同义词替换、文本随机插入、删除等）、时间序列增强（平移、缩放、加噪声等））

数据管理（hugging face、魔搭、scale ai 的nucleus）

模型库和模型列表

数据集库（导入、检索、标签管理）

论坛博客等讨论空间

数据应用（主要是支持下述每一种的取数能力，比如数据配比、交错数据、）

评测：

训练：

知识库



其他的感想两方面的原因最终造成数据工程项目往往虎头蛇尾，开始的时候规划得很大，实际却草草收场

客观上讲，数据建设是一项系统性工程，从组织架构、支撑技术到流程规范，既要有宏观的顶层设计，又要有强有力的落地执行，所以对整个团队的要求会比较高；

主观上讲，本身数据建设经验不足，或者还处于比较初级的阶段，不知道数据建设中有哪些痛点，更不知道用什么样的技术手段和管理机制去解决这些问题。

怎么衡量数据价值

业务部门不可能给你明确的他们要什么数据，他能给你的是他的业务目标是什么。而数据相关的团队，要做的就是对业务目标进行量化，持续跟踪，对于异常要进行诊断分析，给出优化建议，最后一键执行。这个过程最终以数据产品的方式呈现给业务，帮助业务实现数据驱动目标。

数据的价值最终是要回到业务价值上来的，整体价值业务落地价值绑定，比如我们场景下：

可以自动生产数据触发模型迭代占整体的比例来衡量数据的应用价值

提供的数据覆盖度占比

定性价值

解决了什么业务问题，主要通过一些业务场景来体现

帮助在 xxx 指令下风险防控应急时长减少。。

帮助 xxx 领域模型快速迭代。。。

促进生产效率，降低运营成本，优化高价值资产，提升资产的利用率

数据需求的交付时间到底有没有缩短；

还存不存在数据或指标不一致的问题；

数据质量是否有显著的提升；

数据成本是否增长变慢了。