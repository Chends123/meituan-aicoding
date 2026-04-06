<template>
  <div class="soft-card analysis-panel">
    <div class="analysis-panel__header">
      <div>
        <h2>AI 一键分析</h2>
        <p>聚合用户评价，快速提炼问题与亮点</p>
      </div>
      <el-button type="warning" round @click="$emit('analyze')">AI 分析</el-button>
    </div>

    <el-alert
      v-if="streaming"
      title="分析进行中，结果将通过 SSE 逐步返回"
      type="warning"
      :closable="false"
      show-icon
    />

    <div v-if="streaming || analysis.stream_text" class="stream-box">
      <div class="stream-box__title">AI 实时输出</div>
      <pre class="stream-box__content">{{ analysis.stream_text || "正在等待模型返回内容..." }}</pre>
    </div>

    <el-collapse value="summary">
      <el-collapse-item title="整体总结" name="summary">
        <div>{{ analysis.summary || "等待分析结果..." }}</div>
      </el-collapse-item>
    </el-collapse>

    <div class="metric-row">
      <div class="metric-card">
        <span>综合情感评分</span>
        <strong>{{ analysis.sentiment_score || "--" }}</strong>
      </div>
      <div class="metric-card">
        <span>覆盖评价数</span>
        <strong>{{ analysis.review_count || "--" }}</strong>
      </div>
    </div>

    <div class="section-box">
      <h3>正面关键词 Top5</h3>
      <AnalysisKeywordList :items="analysis.positive_keywords || []" />
    </div>

    <div class="section-box">
      <h3>负面关键词 Top5</h3>
      <AnalysisKeywordList :items="analysis.negative_keywords || []" />
    </div>

    <div class="section-box">
      <h3>AI 改进建议</h3>
      <el-timeline v-if="(analysis.suggestions || []).length">
        <el-timeline-item
          v-for="(item, index) in analysis.suggestions"
          :key="index"
          :timestamp="`建议 ${index + 1}`"
        >
          {{ item }}
        </el-timeline-item>
      </el-timeline>
      <div v-else>等待分析结果...</div>
    </div>
  </div>
</template>

<script>
import AnalysisKeywordList from "./AnalysisKeywordList.vue";

export default {
  name: "AIAnalysisPanel",
  components: {
    AnalysisKeywordList
  },
  props: {
    analysis: {
      type: Object,
      default: () => ({})
    },
    streaming: {
      type: Boolean,
      default: false
    }
  }
};
</script>

<style scoped>
.analysis-panel {
  padding: 20px;
}

.analysis-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.analysis-panel__header h2 {
  margin: 0 0 6px;
}

.analysis-panel__header p {
  margin: 0;
  color: var(--mt-subtext);
}

.stream-box {
  margin: 16px 0 18px;
  padding: 14px;
  border-radius: 14px;
  border: 1px solid rgba(255, 195, 0, 0.4);
  background: linear-gradient(180deg, rgba(255, 250, 230, 0.95), rgba(255, 255, 255, 0.95));
}

.stream-box__title {
  font-size: 14px;
  font-weight: 600;
  color: var(--mt-primary-deep);
  margin-bottom: 8px;
}

.stream-box__content {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.6;
  color: var(--mt-text);
  font-family: Consolas, "Courier New", monospace;
  max-height: 220px;
  overflow: auto;
}

.metric-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin: 18px 0;
}

.metric-card {
  border-radius: 14px;
  padding: 16px;
  background: linear-gradient(135deg, rgba(255, 195, 0, 0.18), rgba(255, 255, 255, 0.95));
  border: 1px solid rgba(255, 195, 0, 0.25);
}

.metric-card span {
  display: block;
  color: var(--mt-subtext);
  margin-bottom: 8px;
}

.metric-card strong {
  font-size: 28px;
}

.section-box {
  margin-top: 18px;
}

.section-box h3 {
  margin: 0 0 12px;
}

::v-deep .el-collapse {
  border-color: var(--mt-border);
}

::v-deep .el-collapse-item__header {
  background: transparent;
  color: var(--mt-text);
  border-color: var(--mt-border);
  font-weight: 600;
}

::v-deep .el-collapse-item__header.is-active {
  color: var(--mt-primary-deep);
}

::v-deep .el-collapse-item__arrow {
  color: var(--mt-primary-deep);
}

::v-deep .el-collapse-item__wrap {
  border-color: var(--mt-border);
  background: transparent;
}
</style>
