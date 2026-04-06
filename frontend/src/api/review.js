import http from "./http";

export function fetchReviews(params) {
  return http.get("/reviews", { params });
}

export function fetchReviewTrends(params) {
  return http.get("/reviews/trends", { params });
}
