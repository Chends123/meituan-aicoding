<template>
  <div class="review-card soft-card" :class="{ danger: review.score < 3 }">
    <div class="review-card__header">
      <div>
        <strong>{{ review.username }}</strong>
        <el-tag size="mini" effect="plain">{{ review.score }} 分</el-tag>
        <el-tag v-if="review.score < 3" size="mini" type="danger">低分提醒</el-tag>
      </div>
      <span>{{ formatDate(review.created_at) }}</span>
    </div>
    <p class="review-card__content">{{ review.content }}</p>
  </div>
</template>

<script>
import { formatDateTime } from "../utils/format";

export default {
  name: "ReviewCard",
  props: {
    review: {
      type: Object,
      required: true
    }
  },
  methods: {
    formatDate(value) {
      return formatDateTime(value);
    }
  }
};
</script>

<style scoped>
.review-card {
  padding: 16px;
  margin-bottom: 14px;
}

.review-card.danger {
  background: var(--mt-danger-bg);
  border-color: var(--mt-danger-border);
}

.review-card__header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  margin-bottom: 12px;
  color: var(--mt-subtext);
}

.review-card__header strong {
  color: var(--mt-text);
  margin-right: 8px;
}

.review-card__content {
  margin: 0;
  line-height: 1.7;
  color: var(--mt-text);
}
</style>
