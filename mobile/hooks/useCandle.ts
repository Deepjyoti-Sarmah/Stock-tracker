import { Candle } from "@/types/types";
import { useEffect, useMemo, useState } from "react";

interface Props {
  candles: Candle[];
  visibleChart: "candlesticks" | "line";
}

const TRENDING_COLORS = {
  up: "green",
  down: "red",
  flat: "black"
}

export function useCandle({ candles, visibleChart = "line" }: Props) {
  const newest = candles[candles.length - 1];
  const oldest = candles[0];

  const [trending, setTranding] = useState<"up" | "down" | "flat">("flat")
  const [startToEndDiffetent, setStartToEndDifferent] = useState<{
    amount: Number,
    percentage: Number
  }>({
    amount: 0,
    percentage: 0
  })

  useEffect(() => {
    if (candles.length < 2) return

    const difference = newest.close - oldest.close;
    const percentage = difference / oldest.close * 100;

    setTranding(difference > 0 ? "up" : difference < 0 ? "down" : "flat");
    setStartToEndDifferent({ amount: difference, percentage: percentage });
  }, [candles])

  const chartData = useMemo(() => candles.map(({ timestamp, ...rest }) => ({
    timestamp: new Date(timestamp).getTime(),
    ...(visibleChart === "candlesticks" ? rest : { value: rest.close }),
  })), [candles, visibleChart])

  return {
    trendingColor: TRENDING_COLORS[trending],
    trendingSign: trending === "up" ? "+" : "",
    startToEndDiffetent,
    oldest,
    newest,
    chartData
  }
}
