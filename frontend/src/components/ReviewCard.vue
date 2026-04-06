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

    <div v-if="review.score < 3" class="review-card__actions">
      <el-button size="mini" type="danger" plain :loading="replyStreaming" @click="handleGenerateReply">
        {{ replyText ? "重新生成回复" : "生成商家回复" }}
      </el-button>
    </div>

    <div v-if="replyStreaming || replyText" class="reply-box">
      <div class="reply-box__title">AI 商家回复建议</div>
      <div class="reply-box__content">{{ replyText || "正在生成回复话术..." }}</div>
    </div>
  </div>
</template>

<script>
import { buildReplyStreamURL } from "../api/ai";
import { createEventSource } from "../utils/sse";
import { formatDateTime } from "../utils/format";

export default {
  name: "ReviewCard",
  props: {
    review: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      replyStreaming: false,
      replyText: "",
      replySource: null
    };
  },
  beforeDestroy() {
    this.closeReplySource();
  },
  methods: {
    formatDate(value) {
      return formatDateTime(value);
    },
    handleGenerateReply() {
      if (this.replyStreaming) {
        return;
      }
      this.closeReplySource();
      this.replyStreaming = true;
      this.replyText = "";

      const source = createEventSource(buildReplyStreamURL(this.review.id));
      this.replySource = source;

      source.addEventListener("reply_delta", (event) => {
        const payload = JSON.parse(event.data);
        this.replyText = `${this.replyText}${payload.content || ""}`;
      });

      source.addEventListener("done", (event) => {
        const payload = JSON.parse(event.data || "{}");
        if (!this.replyText && payload.full_content) {
          this.replyText = payload.full_content;
        }
        this.replyStreaming = false;
        this.closeReplySource();
      });

      source.addEventListener("error", (event) => {
        let message = "生成回复建议失败，请稍后重试";
        try {
          const payload = JSON.parse(event.data || "{}");
          if (payload.message) {
            message = payload.message;
          }
        } catch (error) {
          // ignore invalid SSE error payload
        }
        this.replyStreaming = false;
        this.closeReplySource();
        this.$message.error(message);
      });

      source.onerror = () => {
        if (this.replyStreaming) {
          this.replyStreaming = false;
          this.closeReplySource();
          this.$message.error("回复建议连接异常，请稍后重试");
        }
      };
    },
    closeReplySource() {
      if (this.replySource) {
        this.replySource.close();
        this.replySource = null;
      }
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

.review-card__actions {
  margin-top: 14px;
}

.reply-box {
  margin-top: 14px;
  padding: 14px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px dashed rgba(245, 108, 108, 0.45);
}

.reply-box__title {
  font-size: 13px;
  font-weight: 600;
  color: #f56c6c;
  margin-bottom: 8px;
}

.reply-box__content {
  line-height: 1.7;
  color: var(--mt-text);
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
