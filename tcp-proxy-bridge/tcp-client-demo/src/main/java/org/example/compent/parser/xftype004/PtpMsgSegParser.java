package org.example.compent.parser.xftype004;


import lombok.extern.slf4j.Slf4j;
import org.example.models.CommMsgSegModel;
import org.example.models.communication.BunchPlantCommunicationMessageSegment;
import org.example.util.NumUtils;
import org.springframework.stereotype.Component;

import java.util.Date;

/**
 * @Description
 * @Author suoshen
 * @Date 2023/5/18 17:05
 * @Version 1.0
 */

@Component
@Slf4j
public class PtpMsgSegParser implements CommMsgSegParser{

    @Override
    public CommMsgSegModel parse(String binaryData) {

        log.info("点播通信类型-出站信息段解析开始！");
        BunchPlantCommunicationMessageSegment ptp = new BunchPlantCommunicationMessageSegment();
        int count = 0;
        
        String msgTypeBinString =  binaryData.substring(count,count+=3);
        
        String fflxBinString = binaryData.substring(count,count+=1);
        byte fflxByte = Byte.parseByte(fflxBinString,2);
        log.info("发方类型:"+(fflxByte==0?"内网地址":"外网地址"));
        ptp.setSenderType(fflxByte);

        String ssbsBinString = binaryData.substring(count,count+=1);
        byte ssbsByte = Byte.parseByte(ssbsBinString,2);
        log.info("实时标识:"+(ssbsByte==0?"即时推送":"信箱推送"));
        ptp.setActualTime(ssbsByte);

        String sfszBinString = binaryData.substring(count,count+=1);
        byte sfszByte = Byte.parseByte(sfszBinString,2);
        log.info("是否首帧:"+(sfszByte==0?"非首帧":"首帧"));
        ptp.setIfFirstFrame(sfszByte);

        String ywxzBinString = binaryData.substring(count,count+=1);
        byte ywxzByte = Byte.parseByte(ywxzBinString,2);
        log.info("有无续帧:"+(ywxzByte==0?"无续帧":"有续帧"));
        ptp.setIfConsecutiveFrame(ywxzByte);

        String ffdzBinString = "";
        long ffdzLong = 0l;
        if (fflxByte==0){
            ffdzBinString = binaryData.substring(count,count+=24);
            ffdzLong = Long.parseLong(ffdzBinString,2);
        }else if (fflxByte==1){
            ffdzBinString = binaryData.substring(count,count+=48);
            ffdzLong = Long.parseLong(ffdzBinString,2);
        }
        log.info("发方地址:"+ffdzLong);
        ptp.setSenderAddress(ffdzLong);

        String fxsjZzs = binaryData.substring(count,count+=1);
        log.info("发信时间-周指数:"+(fxsjZzs.equals("0")?"本周":"上周"));

        String fxsjZnm = binaryData.substring(count,count+=20);
        int fxsjZnmInt = Integer.parseInt(fxsjZnm,2);
        log.info("发信时间-周内秒:"+fxsjZnmInt);

        Date date = NumUtils.paserBDTime(Long.parseLong(fxsjZzs),fxsjZnmInt);
        log.info("发信时间:"+date);
        ptp.setSenderTime(date);

        String bmlb = binaryData.substring(count,count+=4);
        log.info("编码类别:"+bmlb);
        ptp.setCodingType(Byte.parseByte(bmlb,2));

        String txsjBinString = binaryData.substring(count);
        String txsj = "";
        if ("0000".equals(bmlb)){
            txsj = NumUtils.code2GBK(txsjBinString);
            log.info("数据内容:"+txsjBinString);
            log.info("区位码，数据内容为:"+txsj);
        }else if ("0001".equals(bmlb)){
            log.info("数据内容:"+txsjBinString);
            txsj = NumUtils.paserBCD2String(txsjBinString);
            log.info("BCD码，数据内容为:"+txsj);
        }else {
            log.info("错误的编码类别,内容为:"+txsjBinString);
        }
        ptp.setMessageData(txsj);

        return ptp;
    }

    @Override
    public Boolean support(byte commType) {
        return (commType >> 2) == 0b010;
    }

}
