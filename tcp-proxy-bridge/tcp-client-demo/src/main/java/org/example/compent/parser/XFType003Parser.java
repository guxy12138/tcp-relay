package org.example.compent.parser;

import lombok.extern.slf4j.Slf4j;
import org.example.models.MessageDataModel;
import org.example.util.NumUtils;
import org.springframework.stereotype.Component;

import java.util.Date;

/**
 * @Description
 * @Author suoshen
 * @Date 2023/7/25 11:00
 * @Version 1.0
 */
@Component
@Slf4j
public class XFType003Parser implements XFTypeParser {
    @Override
    public MessageDataModel parse(String data) throws Exception {

        log.info("收到XFType003数据！");

        int count = 0;

        String czlxsbHexString = data.substring(count,count+=2);
        int czlxsbInt = Integer.parseInt(czlxsbHexString,16);
        log.info("出站类型标识:"+czlxsbInt);

        String sflxHexString = data.substring(count,count+=2);
        int sflxInt = Integer.parseInt(sflxHexString,16);
        log.info("收方类型:"+sflxInt);

        String sfdz = "";
        long sfdzLong = 0l;
        if (sflxInt==0){
            log.info("收方类型为0,内网用户");
            sfdz = data.substring(count,count+=6);
            sfdzLong = Long.parseLong(sfdz,16);
        }else if(sflxInt==1){
            log.info("收方类型为1,外网用户");
            sfdz = data.substring(count,count+=12);
            sfdzLong = Long.parseLong(sfdz,16);
        }else {
            return null;
        }
        log.info("收方地址:"+sfdzLong);

        String xxcdHexString = data.substring(count,count+=4);
        int xxcdInt = Integer.parseInt(xxcdHexString,16);
        log.info("信息长度:"+xxcdInt);

        String xxlbHexString = data.substring(count);
        String binaryData = NumUtils.hexString2binaryString(xxlbHexString);
        int binCount = 0;
        log.info("信息类别:"+binaryData);

        String bl = binaryData.substring(binCount,binCount+=1);
        log.info("  ↓保留:"+bl);

        String czzlb = binaryData.substring(binCount,binCount+=2);
        log.info("  ↓出站帧类别:"+czzlb);

        String bs = binaryData.substring(binCount,binCount+=2);
        log.info("  ↓标识:"+bs);

        if ("01".equals(bs)){

            String rnss = binaryData.substring(binCount,binCount+=3);
            log.info("  ↓报告模式：RNSS位置报告:"+rnss);

        }else if ("11".equals(bs)){

            String wzcx = binaryData.substring(binCount,binCount+=3);
            log.info("  ↓报告模式：位置查询请求:"+wzcx);

        }

        String ffdzString = binaryData.substring(binCount,binCount+=24);
        int ffdzInt = Integer.parseInt(ffdzString,2);
        log.info("发方地址:"+ffdzInt);

        if ("01".equals(bs)){
            
            String fxsjZzs = binaryData.substring(binCount,binCount+=1);
            log.info("时间-周指数:"+(fxsjZzs.equals("0")?"本周":"上周"));

            String fxsjZnm = binaryData.substring(binCount,binCount+=20);
            int fxsjZnmInt = Integer.parseInt(fxsjZnm,2);
            log.info("时间-周内秒:"+fxsjZnmInt);

            Date date = NumUtils.paserBDTime(Long.parseLong(fxsjZzs),fxsjZnmInt);
            log.info("时间:"+date);

            String jdfhwString = binaryData.substring(binCount,binCount+=1);
            log.info("经度-符号位:"+(jdfhwString.equals("0")?"东经":"西经"));

            String jdzString = binaryData.substring(binCount,binCount+=23);
            int jdzInt = Integer.parseInt(jdzString,2);
            log.info("经度-经度值:"+jdzInt);

            String wdfhwString = binaryData.substring(binCount,binCount+=1);
            log.info("纬度-符号位:"+(wdfhwString.equals("0")?"北纬":"南纬"));

            String wdzString = binaryData.substring(binCount,binCount+=22);
            int wdzInt = Integer.parseInt(wdzString,2);
            log.info("纬度-经度值:"+wdzInt);

            String gcfhwString = binaryData.substring(binCount,binCount+=1);
            log.info("经度-符号位:"+(gcfhwString.equals("0")?"正":"负"));

            String gczString = binaryData.substring(binCount,binCount+=24);
            int gczInt = Integer.parseInt(gczString,2);
            log.info("经度-经度值:"+gczInt);
        }

        String ztsjString = binaryData.substring(binCount);
        log.info("状态数据:"+ztsjString);

        return null;
    }

    @Override
    public Boolean support(int messageTypeNo) {
        return messageTypeNo == 3;
    }

    public static void main(String[] args) throws Exception {
         new XFType003Parser().parse("01010000006AD13B007E4A6AD13B80000C0000BC00075800000C");
    }
}
