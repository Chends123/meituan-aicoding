const STORAGE_KEY = "meituan-review-ai-analysis-cache";

export function saveAnalysisCache(payload) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(payload));
}

export function loadAnalysisCache() {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw);
  } catch (error) {
    localStorage.removeItem(STORAGE_KEY);
    return null;
  }
}
