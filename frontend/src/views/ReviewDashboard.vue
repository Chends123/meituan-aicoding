<template>
  <div class="page-shell dashboard">
    <section class="dashboard__left">
      <AIAnalysisPanel :analysis="analysis" :streaming="streaming" @analyze="handleAnalyze" />
      <TrendChart :series="trendSeries" />
    </section>

    <section class="dashboard__right">
      <div class="soft-card review-panel">
        <div class="review-panel__header">
          <div>
            <h1>商家评价看板</h1>
            <p>快速查看团购评价、识别风险评论、辅助运营改进</p>
          </div>
        </div>

        <ReviewFilterTabs v-model="activeTab" />
        <ReviewList :reviews="reviews" :hasMore="hasMore" @load-more="loadMore" />
      </div>
    </section>
  </div>
</template>

<script>
import AIAnalysisPanel from "../components/AIAnalysisPanel.vue";
import ReviewFilterTabs from "../components/ReviewFilterTabs.vue";
import ReviewList from "../components/ReviewList.vue";
import TrendChart from "../components/TrendChart.vue";
import { fetchReviews, fetchReviewTrends } from "../api/review";
import { buildAnalysisStreamURL } from "../api/ai";
import { createEventSource } from "../utils/sse";
import { loadAnalysisCache, saveAnalysisCache } from "../utils/storage";

export default {
  name: "ReviewDashboard",
  components: {
    AIAnalysisPanel,
    ReviewFilterTabs,
    ReviewList,
    TrendChart
  },
  data() {
    return {
      activeTab: "all",
      reviews: [],
      hasMore: false,
      nextCursor: null,
      streaming: false,
      trendSeries: [],
      analysis: {
        stream_text: "",
        status_text: "",
        positive_keywords: [],
        negative_keywords: [],
        suggestions: []
      }
    };
  },
  watch: {
    activeTab() {
      this.loadReviews(true);
    }
  },
  created() {
    const cache = loadAnalysisCache();
    if (cache) {
      this.analysis = { ...this.analysis, ...cache, status_text: "" };
    }
    this.loadReviews(true);
    this.loadTrend();
  },
  methods: {
    async loadReviews(reset) {
      if (reset) {
        this.reviews = [];
        this.hasMore = false;
        this.nextCursor = null;
      }
      const params = {
        tab: this.activeTab,
        page_size: 10
      };
      if (!reset && this.nextCursor) {
        params.cursor_time = this.nextCursor.cursor_time;
        params.cursor_id = this.nextCursor.cursor_id;
      }
      try {
        const { data } = await fetchReviews(params);
        this.reviews = reset ? data.list || [] : this.reviews.concat(data.list || []);
        this.hasMore = Boolean(data.has_more);
        this.nextCursor = data.next_cursor || null;
      } catch (error) {
        this.$message.error(error.displayMessage || "获取评价列表失败");
      }
    },
    loadMore() {
      this.loadReviews(false);
    },
    async loadTrend() {
      try {
        const { data } = await fetchReviewTrends({ days: 7 });
        this.trendSeries = data.series || [];
      } catch (error) {
        this.$message.error(error.displayMessage || "获取趋势数据失败");
      }
    },
    handleAnalyze() {
      if (this.streaming) {
        return;
      }
      this.streaming = true;
      this.analysis = {
        stream_text: "",
        status_text: "正在建立分析连接...",
        positive_keywords: [],
        negative_keywords: [],
        suggestions: []
      };
      const source = createEventSource(buildAnalysisStreamURL(this.activeTab));
      source.addEventListener("meta", (event) => {
        const payload = JSON.parse(event.data);
        this.analysis.review_count = payload.review_count;
      });
      source.addEventListener("status", (event) => {
        const payload = JSON.parse(event.data || "{}");
        this.analysis.status_text = payload.message || "正在分析评价内容...";
      });
      source.addEventListener("model_delta", (event) => {
        const payload = JSON.parse(event.data);
        this.analysis.stream_text = `${this.analysis.stream_text || ""}${payload.content || ""}`;
        this.analysis.status_text = "模型正在输出分析内容...";
      });
      source.addEventListener("positive_keywords", (event) => {
        this.analysis.positive_keywords = JSON.parse(event.data).content || [];
      });
      source.addEventListener("negative_keywords", (event) => {
        this.analysis.negative_keywords = JSON.parse(event.data).content || [];
      });
      source.addEventListener("sentiment_score", (event) => {
        this.analysis.sentiment_score = JSON.parse(event.data).content;
      });
      source.addEventListener("suggestions", (event) => {
        this.analysis.suggestions = JSON.parse(event.data).content || [];
      });
      source.addEventListener("summary", (event) => {
        this.analysis.summary = JSON.parse(event.data).content || "";
      });
      source.addEventListener("done", () => {
        this.streaming = false;
        this.analysis.status_text = "";
        saveAnalysisCache(this.analysis);
        source.close();
      });
      source.addEventListener("error", (event) => {
        let message = "AI 分析失败，请稍后重试";
        try {
          const payload = JSON.parse(event.data || "{}");
          if (payload.message) {
            message = payload.message;
          }
        } catch (error) {
          // ignore invalid SSE error payload
        }
        this.streaming = false;
        this.analysis.status_text = "";
        source.close();
        this.$message.error(message);
      });
      source.onerror = () => {
        this.streaming = false;
        this.analysis.status_text = "";
        source.close();
        this.$message.error("AI 分析连接异常，请稍后重试");
      };
    }
  }
};
</script>

<style scoped>
.dashboard {
  display: grid;
  grid-template-columns: 420px 1fr;
  gap: 20px;
  min-height: 100vh;
}

.review-panel {
  padding: 20px;
}

.review-panel__header {
  margin-bottom: 12px;
}

.review-panel__header h1 {
  margin: 0 0 8px;
  font-size: 30px;
}

.review-panel__header p {
  margin: 0;
  color: var(--mt-subtext);
}

@media (max-width: 1100px) {
  .dashboard {
    grid-template-columns: 1fr;
  }
}
</style>
