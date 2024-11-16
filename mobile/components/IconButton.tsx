import { Ionicons } from "@expo/vector-icons";
import { IconProps } from "@expo/vector-icons/build/createIconSet";
import { ComponentProps } from "react";
import { StyleProp, StyleSheet, TouchableOpacity, ViewStyle } from "react-native";

interface Props extends IconProps<ComponentProps<typeof Ionicons>["name"]> {
    onPress: () => void;
    touchableOpicityStyles: StyleProp<ViewStyle>
}

export function IconButton({ touchableOpicityStyles, onPress, ...rest }: Props) {
    return (
        <TouchableOpacity onPress={onPress} style={[styles.touchable, touchableOpicityStyles]}>
            <Ionicons color="white" size={29} {...rest} />
        </TouchableOpacity>
    )
}

const styles = StyleSheet.create({
    touchable: {
        flexDirection: "row",
        borderRadius: 10,
        backgroundColor: "black",
        padding: 10
    }
})