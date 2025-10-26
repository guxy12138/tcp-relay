package org.example.util;

import cn.hutool.core.codec.Base64;

import java.io.ByteArrayInputStream;
import java.io.DataInputStream;
import java.io.IOException;
import java.io.UnsupportedEncodingException;
import java.math.BigInteger;
import java.text.ParseException;
import java.text.SimpleDateFormat;
import java.time.LocalDateTime;
import java.time.ZoneId;
import java.time.format.DateTimeFormatter;
import java.util.Calendar;
import java.util.Date;
import java.util.Locale;
import java.util.TimeZone;

public class NumUtils {
    /**
     * @param str          原始字符串
     * @param formatLength 指定要格式化的长度
     * @return 补0后的字符串
     * @描述: 字符串前面补0
     */
    public static String addZeroForLeft(String str, int formatLength) {
        int strLength = str.length();
        if (formatLength > strLength) {
            // 计算实际需要补0长度
            formatLength -= strLength;
            // 补0操作
            str = String.format("%0" + formatLength + "d", 0) + str;
        }
        return str;
    }


    /**
     * @param number       原始整数
     * @param formatLength 指定要格式化的长度
     * @return 补0后的字符串
     * @描述: 整数前面补0
     */
    public static String addZeroForLeft(int number, int formatLength) {
        // 补0操作
        return String.format("%0" + formatLength + "d", number);
    }

    public static String toBinary(int num) {
        String str = "";
        while (num != 0) {
            str = num % 2 + str;
            num = num / 2;
        }
        return str;
    }

    //将Unicode字符串转换成bool型数组
    public static boolean[] StrToBool(String input) {
        boolean[] output = Binstr16ToBool(BinstrToBinstr16(StrToBinstr(input)));
        return output;
    }

    //将bool型数组转换成Unicode字符串
    public static String BoolToStr(boolean[] input) {
        String output = BinstrToStr(Binstr16ToBinstr(BoolToBinstr16(input)));
        return output;
    }

    //将字符串转换成二进制字符串，以空格相隔
    public static String StrToBinstr(String str) {
        char[] strChar = str.toCharArray();
        String result = "";
        for (int i = 0; i < strChar.length; i++) {
            result += Integer.toBinaryString(strChar[i]) + "";
        }
//        log.info(result.length());
        return result;
    }

    //将二进制字符串转换成Unicode字符串
    public static String BinstrToStr(String binStr) {
        String[] tempStr = StrToStrArray(binStr);
        char[] tempChar = new char[tempStr.length];
        for (int i = 0; i < tempStr.length; i++) {
            tempChar[i] = BinstrToChar(tempStr[i]);
        }
        return String.valueOf(tempChar);
    }

    //将二进制字符串格式化成全16位带空格的Binstr
    public static String BinstrToBinstr16(String input) {
        StringBuffer output = new StringBuffer();
        String[] tempStr = StrToStrArray(input);
        for (int i = 0; i < tempStr.length; i++) {
            for (int j = 16 - tempStr[i].length(); j > 0; j--)
                output.append('0');
            output.append(tempStr[i] + " ");
        }
        return output.toString();
    }

    //将全16位带空格的Binstr转化成去0前缀的带空格Binstr
    public static String Binstr16ToBinstr(String input) {
        StringBuffer output = new StringBuffer();
        String[] tempStr = StrToStrArray(input);
        for (int i = 0; i < tempStr.length; i++) {
            for (int j = 0; j < 16; j++) {
                if (tempStr[i].charAt(j) == '1') {
                    output.append(tempStr[i].substring(j) + " ");
                    break;
                }
                if (j == 15 && tempStr[i].charAt(j) == '0')
                    output.append("0" + " ");
            }
        }
        return output.toString();
    }

    //二进制字串转化为boolean型数组  输入16位有空格的Binstr
    public static boolean[] Binstr16ToBool(String input) {
        String[] tempStr = StrToStrArray(input);
        boolean[] output = new boolean[tempStr.length * 16];
        for (int i = 0, j = 0; i < input.length(); i++, j++)
            if (input.charAt(i) == '1')
                output[j] = true;
            else if (input.charAt(i) == '0')
                output[j] = false;
            else
                j--;
        return output;
    }

    //boolean型数组转化为二进制字串  返回带0前缀16位有空格的Binstr
    public static String BoolToBinstr16(boolean[] input) {
        StringBuffer output = new StringBuffer();
        for (int i = 0; i < input.length; i++) {
            if (input[i])
                output.append('1');
            else
                output.append('0');
            if ((i + 1) % 16 == 0)
                output.append(' ');
        }
        output.append(' ');
        return output.toString();
    }

    //将二进制字符串转换为char
    public static char BinstrToChar(String binStr) {
        int[] temp = BinstrToIntArray(binStr);
        int sum = 0;
        for (int i = 0; i < temp.length; i++) {
            sum += temp[temp.length - 1 - i] << i;
        }
        return (char) sum;
    }

    //将初始二进制字符串转换成字符串数组，以空格相隔
    public static String[] StrToStrArray(String str) {
        return str.split(" ");
    }

    //将二进制字符串转换成int数组
    public static int[] BinstrToIntArray(String binStr) {
        char[] temp = binStr.toCharArray();
        int[] result = new int[temp.length];
        for (int i = 0; i < temp.length; i++) {
            result[i] = temp[i] - 48;
        }
        return result;
    }

    /**
     * 数组转换为16进制字符串
     *
     * @param src
     * @return
     */
    public static String bytesToHexString(byte[] src) {
        StringBuilder stringBuilder = new StringBuilder("");
        if (src == null || src.length <= 0) {
            return null;
        }
        for (int i = 0; i < src.length; i++) {
            int v = src[i] & 0xFF;
            String hv = Integer.toHexString(v);
            if (hv.length() < 2) {
                stringBuilder.append(0);
            }
            stringBuilder.append(hv);
        }
        return stringBuilder.toString();
    }
    
    private static byte charToByte(char c) {
        return (byte) "0123456789ABCDEF".indexOf(c);
    }

    /**
     * hex字符串转byte数组
     * @param hexString
     * @return
     */
    public static byte[] hexStringToBytes(String hexString) {
        if (hexString == null || hexString.equals("")) {
            return null;
        }
        hexString = hexString.toUpperCase();
        int length = hexString.length() / 2;
        char[] hexChars = hexString.toCharArray();
        byte[] d = new byte[length];
        for (int i = 0; i < length; i++) {
            int pos = i * 2;
            d[i] = (byte) (charToByte(hexChars[pos]) << 4 | charToByte(hexChars[pos + 1]));
        }
        return d;
    }


    /**
     * hex字符串转byte数组
     *
     * @param inHex 待转换的Hex字符串
     * @return 转换后的byte数组结果
     */
    public static byte[] hexToByteArray(String inHex, int len) {
        int hexlen = inHex.length();
        int hLen = hexlen / 2;
        byte[] result = new byte[(len)];
        if (hexlen % 2 == 1) {
            hLen += 1;
            //奇数
            hexlen++;
            inHex = "0" + inHex;
        }
        int j = 0;
        for (int k = 0; k < len - hLen; k++) {
            result[j] = 0 & 0xff;
            j++;
        }
        for (int i = 0; i < hexlen; i += 2) {
            result[j] = (byte) Integer.parseInt(inHex.substring(i, i + 2), 16);
            j++;
        }
        return result;
    }


    /**
     * 16进制转换成为string类型字符串
     *
     * @param s
     * @return
     */
    public static String hexStringToString(String s) {
        if (s == null || s.equals("")) {
            return null;
        }
        s = s.replace(" ", "");
        byte[] baKeyword = new byte[s.length() / 2];
        for (int i = 0; i < baKeyword.length; i++) {
            try {
                baKeyword[i] = (byte) (0xff & Integer.parseInt(
                        s.substring(i * 2, i * 2 + 2), 16));
            } catch (Exception e) {
                e.printStackTrace();
            }
        }
        try {
            s = new String(baKeyword, "UTF-8");
            new String();
        } catch (Exception e1) {
            e1.printStackTrace();
        }
        return s;
    }

    /**
     * 截取指定数组  不动原始数组
     *
     * @param src
     * @param begin
     * @param count
     * @return
     */
    public static byte[] subBytes(byte[] src, int begin, int count) {
        byte[] bs = new byte[count];
        System.arraycopy(src, begin, bs, 0, count);
        return bs;
    }

    /**
     * 截取指定数组-原始数组改变
     *
     * @param dis
     * @param count
     * @return
     */
    public static byte[] subBytes(DataInputStream dis, int count) throws IOException {
        byte[] bs = new byte[count];
        for (int i = 0; i < count; i++) {
            bs[i]=dis.readByte();
        }
        return bs;
    }


    /**
     * byte数组中取int数值，本方法适用于(低位在后，高位在前)的顺序。和intToBytes2（）配套使用
     */
    public static long bytesToLong(byte[] src, int offset, int n) {
        long j = 0;
        for (int i = 0; i < n; i++) {
            j = j | (src[offset + i] & 0xFF) << (8 * i);
        }
        return j;
    }


    /**
     * 字符串转换成为16进制(无需Unicode编码)
     *
     * @param str
     * @return
     */
    public static String str2HexStr(String str) {
        char[] chars = "0123456789ABCDEF".toCharArray();
        StringBuilder sb = new StringBuilder("");
        byte[] bs = str.getBytes();
        int bit;
        for (int i = 0; i < bs.length; i++) {
            bit = (bs[i] & 0x0f0) >> 4;
            sb.append(chars[bit]);
            bit = bs[i] & 0x0f;
            sb.append(chars[bit]);
            // sb.append(' ');
        }
        return sb.toString().trim();
    }

    /**
     * 16进制直接转换成为字符串(无需Unicode解码)
     *
     * @param hexStr
     * @return
     */
    public static String hexStr2Str(String hexStr) {
        String str = "0123456789ABCDEF";
        char[] hexs = hexStr.toCharArray();
        byte[] bytes = new byte[hexStr.length() / 2];
        int n;
        for (int i = 0; i < bytes.length; i++) {
            n = str.indexOf(hexs[2 * i]) * 16;
            n += str.indexOf(hexs[2 * i + 1]);
            bytes[i] = (byte) (n & 0xff);
        }
        return new String(bytes);
    }





    /**
     * 将二进制数组转为任意进制字符串
     * @param bytes 二进制数组
     * @param radix 需要转换成的进制
     * @return
     */
    public static String binary(byte[] bytes, int radix){
        String s = new BigInteger(1, bytes).toString(radix);// 这里的1代表正数
        int a=s.length() % 8;

        int b=0;
        if(a!=0){
            b=8-a;
        }

        for(int n=0;n<b;n++){
            s="0"+s;
        }
        return s;
    }

    /**
     * 格式化日期,将毫秒数转化为日期
     * @param time 日期毫秒数
     * @return
     */
    public static Date formatDate(long time) {
        Date returnDate=null;
        try {
            Date date=new Date(time);
            SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");

            String formatStr=sdf.format(date);
            returnDate =sdf.parse(formatStr);
        }catch (Exception e){
            e.printStackTrace();
        }
        return  returnDate;
    }


    /**
     * 二制度字符串转字节数组，如 101000000100100101110000 -> A0 09 70
     * @param input 输入字符串。
     * @return 转换好的字节数组。
     */
    public static byte[] string2bytes(String input) {
        StringBuilder in = new StringBuilder(input);
        // 注：这里in.length() 不可在for循环内调用，因为长度在变化
        int remainder = in.length() % 8;
        if (remainder > 0)
            for (int i = 0; i < 8 - remainder; i++)
                in.append("0");
        byte[] bts = new byte[in.length() / 8];

        // Step 8 Apply compression
        for (int i = 0; i < bts.length; i++)
            bts[i] = (byte) Integer.parseInt(in.substring(i * 8, i * 8 + 8), 2);

        return bts;
    }


    /**
     * 16进制字符串转2进制字符串
     * @param hexString
     * @return
     */
    public static String hexString2binaryString(String hexString) {
        if (hexString == null || hexString.length() % 2 != 0)
            return null;
        String bString = "", tmp;
        for (int i = 0; i < hexString.length(); i++) {
            tmp = "0000" + Integer.toBinaryString(Integer.parseInt(hexString.substring(i, i + 1), 16));
            bString += tmp.substring(tmp.length() - 4);
        }
        return bString;
    }

    /**
     * 2进制字符串转16进制字符串
     * @param bString
     * @return
     */
    public static String binaryString2hexString(String bString) {
        if (bString == null || bString.equals("") || bString.length() % 8 != 0)
            return null;
        StringBuffer tmp=new StringBuffer();
        int iTmp = 0;
        for (int i = 0; i < bString.length(); i += 4) {
            iTmp = 0;
            for (int j = 0; j < 4; j++) {
                iTmp += Integer.parseInt(bString.substring(i + j, i + j + 1)) << (4 - j - 1);
            }
            tmp.append(Integer.toHexString(iTmp));
        }
        return tmp.toString();
    }


    /**
     * 秒数转日期
     * @param time
     * @return
     */
    public static Date paserTime(long time){
        Date parse=null;
        System.setProperty("user.timezone", "UTC");
        TimeZone tz = TimeZone.getTimeZone("UTC");
        TimeZone.setDefault(tz);
        SimpleDateFormat format = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
        String times = format.format(new Date(time * 1000L));
        try {
            parse = format.parse(times);
        } catch (ParseException e) {
            e.printStackTrace();
        }

        return parse;
    }


    /**
     * 将周指示周内秒转换为时间格式 yyyy-MM-dd
     * @param zzs 周指示
     * @param znm 周内秒
     * @return yyyy-MM-dd 北斗时
     */
    public static String paserBDTime(int zzs,int znm) {

        Calendar calendar = Calendar.getInstance(TimeZone.getTimeZone("UTC"));

        if (zzs==1){
            calendar.add(Calendar.WEEK_OF_YEAR,-1);
        }

        calendar.set(Calendar.DAY_OF_WEEK,calendar.getActualMinimum(Calendar.DAY_OF_WEEK));
        calendar.set(Calendar.HOUR_OF_DAY,calendar.getActualMinimum(Calendar.HOUR_OF_DAY));
        calendar.set(Calendar.MINUTE,calendar.getActualMinimum(Calendar.MINUTE));
        calendar.set(Calendar.SECOND,calendar.getActualMinimum(Calendar.SECOND));
        calendar.set(Calendar.MILLISECOND,calendar.getActualMinimum(Calendar.MILLISECOND));
        calendar.add(Calendar.SECOND,znm);

        LocalDateTime localDateTime = LocalDateTime.ofInstant(calendar.getTime().toInstant(), ZoneId.systemDefault());
        String formatStr = localDateTime.format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss.SSS"));

        return  formatStr;
    }

    /**
     * 将周指示周内秒转换为时间格式 Date
     * @param zzs 周指示
     * @param znm 周内秒
     * @return Date utc时
     */
    public static Date paserBDTime(Long zzs,int znm) {

        Calendar calendar = Calendar.getInstance(TimeZone.getTimeZone("UTC"));

        if (zzs==1){
            calendar.add(Calendar.WEEK_OF_YEAR,-1);
        }

        calendar.set(Calendar.DAY_OF_WEEK,calendar.getActualMinimum(Calendar.DAY_OF_WEEK));
        calendar.set(Calendar.HOUR_OF_DAY,calendar.getActualMinimum(Calendar.HOUR_OF_DAY));
        calendar.set(Calendar.MINUTE,calendar.getActualMinimum(Calendar.MINUTE));
        calendar.set(Calendar.SECOND,calendar.getActualMinimum(Calendar.SECOND));
        calendar.set(Calendar.MILLISECOND,calendar.getActualMinimum(Calendar.MILLISECOND));
        calendar.add(Calendar.SECOND,znm);

        return  calendar.getTime();
    }

    /**
     *  bcd码转十进制字符串
     * @param bcd
     * @return
     */
    public static String paserBCD2String(String bcd){
        for (int i = bcd.length()%4 ; i > 0 ; i-- ){
            bcd += "0";
        }
        int length = bcd.length()/4;
        byte b[] = new byte[length];
        String returnString = "";
        for (int i = 0; i < b.length; i++) {
            b[i] = Byte.parseByte(bcd.substring((i)*4,(i+1)*4),2);
            returnString += b[i]+"";
        }
        // 2+2+4+1+2+1+1+2+2+1+1+2
        return returnString;
    }

    /**
     *  区位码转机内码
     * @param binary
     * @return
     */
    public static String code2GBK(String binary)  {
        StringBuffer sb = new StringBuffer(binary.replaceAll(" ","").trim());
        System.out.println(sb);
        String returnString = "";
        int bytesLength = sb.length() / 7;
        for (int i = 0; i < bytesLength; i++) {
            sb.insert(i*8,"0");
        }
        byte gbkBinaryBytes[] = NumUtils.string2bytes(sb.toString());
        for (int i = 0; i < gbkBinaryBytes.length; i++) {
            gbkBinaryBytes[i] += 0xA0;
        }
        try {
            returnString = new String(gbkBinaryBytes,"gb2312");
        }catch (UnsupportedEncodingException e){
            e.printStackTrace();
        }
        return returnString;
    }

    public static String code2HB(String code,int radix) throws UnsupportedEncodingException {
        String hexString = "";
        if (radix==2){
            hexString = NumUtils.binaryString2hexString(code);
        }else if (radix==4){
            hexString = code;
        }
        if ("A4".equalsIgnoreCase(hexString.replaceAll(" ","").trim().substring(0,2))){
            StringBuffer sb = new StringBuffer(hexString.replaceAll(" ","").trim().substring(2));
            String result = new String(NumUtils.hexStringToBytes(sb.toString()),"gb2312");
            return result;
        }else {
            throw (new RuntimeException("开头字节不为0xA4，此字段不是混编！"));
        }
    }

    public static void main(String[] args) throws UnsupportedEncodingException {
        //3690 0b0010 0100 0101 1010 您 2635 0b0001 1010 0010 0011 好 4650 0b0010 1110 0011 0010 我 4239 0b0010 1010 0010 0111 是 1866 0b0001 0010 0100 0010 测 4252 0b0010 1010 0011 0100 试 4293 0b0010 1010 0101 1101 数 3061 0b0001 1110 0011 1101 据 5027 0b0011 0010 0001 1011 一
//        System.out.println(NumUtils.paserBDTime(0,180229));
//        System.out.println(NumUtils.code2GBK("010 0100101 1010 001 1010 010 0011 000 0011 000 1100 010 1110 011 0010 010 1010010 0111 001 0010100 0010010 1010011 0100010 1010 101 1101001 1110 011 1101011 0010 001 1011"));
//        String a = "39 36 62 65 31 62 35 33 65 33 63 34 33 62 62 30 31 31 33 35 63 33 32 66 66 37 61 31 37 32 39 62";
//        System.out.println(new String(NumUtils.hexStringToBytes(a.replaceAll(" ","")),"ascii"));
//        int b  = 1 ;
//        System.out.println((double) b * 1 );
//        System.out.println(NumUtils.code2HB("A4 7E 02 00 00 1C 01 38 80 00 00 10 00 07 00 00 00 00 00 00 00 01 01 61 89 DC 06 C2 AF 58 00 00 00 00 00 00 00 00 06 08 00 00 B9 7E",4));
        System.out.println(NumUtils.code2GBK("0010001"));

    }
}
