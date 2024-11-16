import { Candle } from "@/types/types";
import { StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { StockImage } from "./StockImage";
import { useCandle } from "@/hooks/useCandle";
import { LineChart, TLineChartContext, TLineChartDataProp } from "react-native-wagmi-charts"

interface Props {
    symbol: string;
    candles: Candle[];
    onPress: () => void;
}

export function StockRow({ candles, symbol, onPress }: Props) {

    const {
        chartData,
        newest,
        trendingColor,
        trendingSign,
        startToEndDiffetent
    } = useCandle({ candles })

    return (
        <TouchableOpacity style={styles.container} onPress={onPress}>
            <StockImage style={styles.img} symbol={symbol} />
            <View style={styles.imageContainer}>
                <Text style={styles.symbol}></Text>
            </View>
            <LineChart.Provider data={chartData as TLineChartDataProp} >
                <LineChart width={100} height={100}>
                    <LineChart.Path color={trendingColor}>
                        <LineChart.Gradient />
                        <LineChart.HorizontalLine color={trendingColor} at={{ index: 0 }} />
                    </LineChart.Path>
                </LineChart>
            </LineChart.Provider>

            <View style={styles.priceContainer} >
                <Text style={styles.price}>
                    {"$ " + newest.close.toFixed(2)}
                </Text>
                <Text
                    style={[styles.priceStatus, { color: trendingColor }]}
                >
                    {trendingSign}
                    {startToEndDiffetent.amount.toFixed(2)}
                    {" "}
                    ({trendingSign}{startToEndDiffetent.percentage.toFixed(2) + "%"})
                </Text>
            </View>
        </TouchableOpacity>
    )
}

const styles = StyleSheet.create({
    container: {
        flexDirection: "row",
        justifyContent: "space-between",
        alignItems: "center",
        paddingHorizontal: 10,
    },
    imageContainer: {
        flexDirection: "row",
        justifyContent: "center",
        alignItems: "center",
        gap: 10
    },
    img: {
        width: 60,
        height: 60,
        margin: 10,
        borderRadius: 30,
    },
    symbol: {
        fontSize: 18,
        fontWeight: "bold",
    },
    priceContainer: {
        justifyContent: "center",
        alignItems: "flex-end",
        gap: 5,
    },
    price: {
        fontSize: 22,
        fontWeight: "bold"
    },
    priceStatus: {
        fontSize: 15,
        fontWeight: "semibold"
    }
})