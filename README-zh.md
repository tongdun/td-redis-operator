
![td-redis-operator](docs/imgs/td-redis-operator-logo.jpg)


Switch Languages: <a href="README.md">English Documents</a> | <a href="README-zh.md">中文文档</a>

<br>

<div class="blog_content">
    <div class="iteye-blog-content-contain" style="font-size: 14px;">
<h1>项目背景</h1>
<p style="font-size: 14px;">
&nbsp;&nbsp;&nbsp;同盾做为国内头部第三方风控公司，日均处理决策请求高达百亿次。因此在同盾的主体数据存储基础架构中，大量使用Redis做为缓存组件。在业务高峰时期，集群实际部署高达千余Redis实例，这势必对DBA运维管控带来极大挑战。2018年，集团推动无状态应用全面容器化，结合云原生技术的缓存云产品开始在数据存储和云原生团队内部酝酿落地 </p>
<p style="font-size: 14px;">
&nbsp;&nbsp;&nbsp;td-redis-operator第一版本可追溯到2018年，此次外部开源的版本为第2版，开发时间从2018年7月份一直持续到现在，目前同盾两地双中心的Redis集群全部部署在超大规模的Kubernates上。
</p>
<strong>目前使用规模：</strong>
<ol style="font-size: 14px;">
<li>Redis实例2000+</li>
<li>PB级别数据量量</li>
<li>涉及200+个在线实时业务</li>
</ol>

<h1>项目介绍</h1>
<p style="margin-top: 15px; margin-bottom: 15px; color: #333333; font-family: Helvetica, arial, freesans, clean, sans-serif; font-size: 15px; line-height: 25px;">名称：td-redis-operator</p>
<p style="margin-top: 15px; margin-bottom: 15px; color: #333333; font-family: Helvetica, arial, freesans, clean, sans-serif; font-size: 15px; line-height: 25px;">语言： 纯go开发</p>
<p style="margin-top: 15px; margin-bottom: 15px; color: #333333; font-family: Helvetica, arial, freesans, clean, sans-serif; font-size: 15px; line-height: 25px;">定位： 完全基于云原生技术，实现资源生命周期管理、故障自愈、HA等</p>
<p> </p>
<h1>工作原理</h1>
<p><img width="848" src="https://github.com/tongdun/td-redis-operator/blob/gaoshengL-patch-1/1.png" height="500" alt=""></p>
<p>原理描述：</p>
<p>1.   基于Operator开源产品，完全在Kubernate上运维托管。 什么是Kubernate Operator,  请<a href="https://kubernetes.io/docs/concepts/extend-kubernetes/operator/">点击</a></p>
<p>2.   支持两种Redis实例管理交付，即Redis主备和RedisCluster</p>
<p> </p>
<h1>Introduction</h1>
<p>See the page for Introduction: [[Introduction]].</p>
<h1>QuickStart</h1>
<p>See the page for quick start: [[QuickStart]].</p>
<p> </p>
<h1>AdminGuide</h1>
<p>See the page for admin deploy guide : [[AdminGuide]]</p>
<p> </p>
<h1>常见问题</h1>
<p>See the page for FAQ: [[FAQ]]</p>
<p> </p>
<p><span style="color: #333333; font-family: Helvetica, arial, freesans, clean, sans-serif; font-size: 15px; line-height: 25px;">2.  </span><span style="color: #333333; font-family: Helvetica, arial, freesans, clean, sans-serif; font-size: 15px; line-height: 25px;">邮件交流： gaosheng.liang1024@gmail.com</span></p>
<p><span style="color: #333333; font-family: Helvetica, arial, freesans, clean, sans-serif; font-size: 15px; line-height: 25px;">4.  </span><span style="color: #333333; font-family: Helvetica, arial, freesans, clean, sans-serif; font-size: 15px; line-height: 25px;">报告issue：</span><a href="https://github.com/tongdun/td-redis-operator/issues" style="color: #4183c4; font-family: Helvetica, arial, freesans, clean, sans-serif; font-size: 15px; line-height: 25px;">issues</a></p>
<p> </p>
</div>
