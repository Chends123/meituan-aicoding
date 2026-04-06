<template>
  <div class="soft-card trend-chart">
    <div class="chart-title">近 7 日评分趋势</div>
    <div ref="chart" class="chart-body"></div>
  </div>
</template>

<script>
import * as echarts from "echarts";

export default {
  name: "TrendChart",
  props: {
    series: {
      type: Array,
      default: () => []
    }
  },
  watch: {
    series: {
      deep: true,
      handler() {
        this.renderChart();
      }
    }
  },
  mounted() {
    this.chart = echarts.init(this.$refs.chart);
    this.renderChart();
    window.addEventListener("resize", this.handleResize);
  },
  beforeDestroy() {
    window.removeEventListener("resize", this.handleResize);
    if (this.chart) {
      this.chart.dispose();
    }
  },
  methods: {
    handleResize() {
      if (this.chart) {
        this.chart.resize();
      }
    },
    renderChart() {
      if (!this.chart) {
        return;
      }
      const dates = this.series.map((item) => item.date);
      const avgScores = this.series.map((item) => item.avg_score);
      const reviewCounts = this.series.map((item) => item.review_count);
      this.chart.setOption({
        tooltip: { trigger: "axis" },
        legend: { data: ["平均评分", "评论数"] },
        grid: { top: 48, left: 42, right: 24, bottom: 36 },
        xAxis: { type: "category", data: dates },
        yAxis: [
          { type: "value", min: 0, max: 5, name: "评分" },
          { type: "value", minInterval: 1, name: "评论数" }
        ],
        series: [
          {
            name: "平均评分",
            type: "line",
            smooth: true,
            data: avgScores,
            lineStyle: { color: "#ffb000", width: 3 },
            itemStyle: { color: "#ffb000" }
          },
          {
            name: "评论数",
            type: "bar",
            yAxisIndex: 1,
            data: reviewCounts,
            itemStyle: { color: "rgba(255, 195, 0, 0.45)" }
          }
        ]
      });
    }
  }
};
</script>

<style scoped>
.trend-chart {
  margin-top: 20px;
  padding: 16px;
}

.chart-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 12px;
}

.chart-body {
  height: 320px;
}
</style>
