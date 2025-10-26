package org.example.util;

import org.springframework.lang.Nullable;
import org.springframework.util.StringUtils;

import java.math.BigInteger;
import java.nio.charset.StandardCharsets;
import java.util.stream.IntStream;

public class BinaryTransUtils {
    /**
     * 将十六进制的字符串转换成二进制的字符串
     *
     * @param hexString
     * @return
     */
    public static String hexStrToBinaryStr(String hexString) {

        if (hexString == null || hexString.equals("")) {
            return null;
        }
        StringBuffer sb = new StringBuffer();
        // 将每一个十六进制字符分别转换成一个四位的二进制字符
        for (int i = 0; i < hexString.length(); i++) {
            String indexStr = hexString.substring(i, i + 1);
            String binaryStr = Integer.toBinaryString(Integer.parseInt(indexStr, 16));
            while (binaryStr.length() < 4) {
                binaryStr = "0" + binaryStr;
            }
            sb.append(binaryStr);
        }

        return sb.toString();
    }

    public static String intToBinaryString(int original, int digitCapacity) {
        //将int转化为2进制串
        String binary = Integer.toBinaryString(original);
        int currentlength = binary.length();
        if (digitCapacity - currentlength >= 1) {
            //需要不位
            StringBuilder stringBuilder = new StringBuilder();
            //补位数
            IntStream.range(0, digitCapacity - currentlength).forEach(i -> stringBuilder.append("0"));
            binary = stringBuilder + binary;
        }
        return binary;
    }

    public static String longToBinaryString(Long original, int digitCapacity) {
        //将Long转化为2进制串
        String binary = Long.toBinaryString(original);
        int currentlength = binary.length();
        if (digitCapacity - currentlength >= 1) {
            //需要不位
            StringBuilder stringBuilder = new StringBuilder();
            //补位数
            IntStream.range(0, digitCapacity - currentlength).forEach(i -> stringBuilder.append("0"));
            binary = stringBuilder + binary;
        }
        return binary;
    }

    public static byte[] string2bytes(String input) {
        StringBuilder in = new StringBuilder(input);
        // 注：这里in.length() 不可在for循环内调用，因为长度在变化
        int remainder = in.length() % 8;
        if (remainder > 0) for (int i = 0; i < 8 - remainder; i++)
            in.append("0");
        byte[] bts = new byte[in.length() / 8];

        for (int i = 0; i < bts.length; i++)
            bts[i] = (byte) Integer.parseInt(in.substring(i * 8, i * 8 + 8), 2);
        return bts;
    }
//11000100
//    byte数组转换为二进制字符串

    public static String byteArrToBinStr(byte[] b) {
        StringBuilder result = new StringBuilder();
        for (byte value : b) {
            result.append(Long.toString(value & 0xff, 2));
        }
        return result.toString().substring(0, result.length() - 1);
    }

    //将bits转化为各种进制的字符串
    public static String binary(byte[] bytes, int radix) {
        return new BigInteger(1, bytes).toString(radix);
    }

    public static String bin2hex(String input) {
        StringBuilder sb = new StringBuilder();
        int len = input.length();
        System.out.println("原数据长度：" + (len / 8) + "字节");

        for (int i = 0; i < len / 4; i++) {
            //每4个二进制位转换为1个十六进制位
            String temp = input.substring(i * 4, (i + 1) * 4);
            int tempInt = Integer.parseInt(temp, 2);
            String tempHex = Integer.toHexString(tempInt).toUpperCase();
            sb.append(tempHex);
        }

        return sb.toString();
    }
//    public static void main(String[] args) {
//        String token = "000000110010001000000000000101000000000000000000000000000000000100000001000000000010100000000000000000000000000000000000000001111110011100000011000111000000000001100100000000000010000000110001001100010011000100110001001100010011000100110001001100010011000100110001001100010011000100110001001100010011000100110001001100010011000100110001001100010011000100110001001100010011000100110001001100010011000100110001001100010011000100110001001100010111100001111000011110000111100010001000100010001000100010001000";
//        bin2hex(token);
//        System.out.println("ok");
//    }
}
