<template>
  <div>
    <ReviewCard v-for="item in reviews" :key="item.id" :review="item" />
    <div v-if="!reviews.length" class="empty-box soft-card">暂无评价数据</div>
    <div v-if="hasMore" class="load-more">
      <el-button round @click="$emit('load-more')">加载更多</el-button>
    </div>
  </div>
</template>

<script>
import ReviewCard from "./ReviewCard.vue";

export default {
  name: "ReviewList",
  components: {
    ReviewCard
  },
  props: {
    reviews: {
      type: Array,
      default: () => []
    },
    hasMore: {
      type: Boolean,
      default: false
    }
  }
};
</script>

<style scoped>
.empty-box {
  padding: 28px;
  text-align: center;
  color: var(--mt-subtext);
}

.load-more {
  text-align: center;
  padding: 12px 0 4px;
}
</style>
