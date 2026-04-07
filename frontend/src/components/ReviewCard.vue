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
      <div class="reply-box__title">
        <span>AI 商家回复建议</span>
        <span v-if="replyStreaming" class="reply-box__status">{{ replyStatusText || '模型正在输出回复建议...' }}</span>
      </div>
      <div ref="replyBox" class="reply-box__content">{{ replyText || replyStatusText || '正在生成回复话术...' }}<span v-if="replyStreaming" class="stream-caret">|</span></div>
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
      replyStatusText: "",
      replySource: null
    };
  },
  beforeDestroy() {
    this.closeReplySource();
  },
  updated() {
    const target = this.$refs.replyBox;
    if (target) {
      target.scrollTop = target.scrollHeight;
    }
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
      this.replyStatusText = "正在建立回复连接...";

      const source = createEventSource(buildReplyStreamURL(this.review.id));
      this.replySource = source;

      source.addEventListener("status", (event) => {
        const payload = JSON.parse(event.data || "{}");
        this.replyStatusText = payload.message || "正在生成回复建议...";
      });

      source.addEventListener("reply_delta", (event) => {
        const payload = JSON.parse(event.data);
        this.replyText = `${this.replyText}${payload.content || ""}`;
        this.replyStatusText = "模型正在输出回复内容...";
      });

      source.addEventListener("done", (event) => {
        const payload = JSON.parse(event.data || "{}");
        if (!this.replyText && payload.full_content) {
          this.replyText = payload.full_content;
        }
        this.replyStatusText = "";
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
        this.replyStatusText = "";
        this.replyStreaming = false;
        this.closeReplySource();
        this.$message.error(message);
      });

      source.onerror = () => {
        if (this.replyStreaming) {
          this.replyStatusText = "";
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
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
  font-weight: 600;
  color: #f56c6c;
  margin-bottom: 8px;
}

.reply-box__status {
  font-size: 12px;
  color: var(--mt-subtext);
}

.reply-box__content {
  line-height: 1.7;
  color: var(--mt-text);
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 220px;
  overflow: auto;
}

.stream-caret {
  display: inline-block;
  margin-left: 2px;
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%,
  50% {
    opacity: 1;
  }
  51%,
  100% {
    opacity: 0;
  }
}
</style>
