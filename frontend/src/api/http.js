import axios from "axios";

const http = axios.create({
  baseURL: "/api/v1",
  timeout: 15000
});

http.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status;
    const serverMsg = error.response?.data?.message || error.response?.data?.msg;
    if (serverMsg) {
      error.displayMessage = serverMsg;
    } else if (status === 500) {
      error.displayMessage = "服务器内部错误，请稍后重试";
    } else if (status === 404) {
      error.displayMessage = "请求的资源不存在";
    } else if (error.code === "ECONNABORTED") {
      error.displayMessage = "请求超时，请检查网络连接";
    } else {
      error.displayMessage = "网络异常，请检查网络连接";
    }
    return Promise.reject(error);
  }
);

export default http;
