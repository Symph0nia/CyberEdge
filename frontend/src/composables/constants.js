// constants.js
export const HTTP_STATUS_CLASSES = {
  "2xx": "bg-green-500/20 text-green-300",
  "3xx": "bg-blue-500/20 text-blue-300",
  "4xx": "bg-yellow-500/20 text-yellow-300",
  "5xx": "bg-red-500/20 text-red-300",
  default: "bg-gray-500/20 text-gray-300",
};

export const getHttpStatusClass = (status) => {
  if (status >= 200 && status < 300) return HTTP_STATUS_CLASSES["2xx"];
  if (status >= 300 && status < 400) return HTTP_STATUS_CLASSES["3xx"];
  if (status >= 400 && status < 500) return HTTP_STATUS_CLASSES["4xx"];
  if (status >= 500) return HTTP_STATUS_CLASSES["5xx"];
  return HTTP_STATUS_CLASSES["default"];
};
