<template>
  <div class="keyword-list">
    <el-tag v-for="(item, index) in items" :key="index" effect="plain" class="keyword-tag">
      {{ typeof item === "string" ? item : item.keyword }}
    </el-tag>
  </div>
</template>

<script>
export default {
  name: "AnalysisKeywordList",
  props: {
    items: {
      type: Array,
      default: () => []
    }
  }
};
</script>

<style scoped>
.keyword-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.keyword-tag {
  margin: 0;
}
</style>
