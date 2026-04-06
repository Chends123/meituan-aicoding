export function buildAnalysisStreamURL(tab) {
  const search = new URLSearchParams();
  search.set("tab", tab || "all");
  return `/api/v1/ai/reviews/analyze/stream?${search.toString()}`;
}

export function buildReplyStreamURL(reviewId) {
  return `/api/v1/ai/reviews/${reviewId}/reply/stream`;
}
