import { Platform } from "react-native";

export const baseUrl = (scheme: "http" | "ws") => {
  const PORT = 3000;
  const HOST = Platform.OS === "android" ? "192.168.165.165" : "localhost";
  //    const HOST = Platform.OS === "android" ? "172.19.0.1 " "10.0.2.2" : "localhost";

  return `${scheme}://${HOST}:${PORT}`;
}
