package org.example.compent.parser;

import lombok.extern.slf4j.Slf4j;
import org.example.compent.parser.xftype004.CommMsgSegParser;
import org.example.models.CommMsgSegModel;
import org.example.models.MessageDataModel;

import org.example.models.xftype.XFType004;
import org.example.util.NumUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * @Description
 * @Author suoshen
 * @Date 2023/5/18 13:12
 * @Version 1.0
 */
@Component
@Slf4j
public class XFType004Parser implements XFTypeParser{

    @Autowired
    private List<CommMsgSegParser> parsers;

    @Override
    public MessageDataModel parse(String data) {

        log.info("收到XFType004数据！");

        int count = 0;
        XFType004 xfType004 = new XFType004();

        String czlxsbHexString = data.substring(count,count+=2);
        int czlxsbInt = Integer.parseInt(czlxsbHexString,16);
        log.info("出站类型标识:"+czlxsbInt);
        xfType004.setDeparture(czlxsbInt);

        String sflxHexString = data.substring(count,count+=2);
        int sflxInt = Integer.parseInt(sflxHexString,16);
        log.info("收方类型:"+sflxInt);
        xfType004.setRecieverType(sflxInt);

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
        xfType004.setRecieverAddress(sfdzLong);

        String xxcdHexString = data.substring(count,count+=4);
        int xxcdInt = Integer.parseInt(xxcdHexString,16);
        log.info("信息长度:"+xxcdInt);
        xfType004.setMessageLength(xxcdInt);

        String xxlbHexString = data.substring(count);
        String xxlbBinString = NumUtils.hexString2binaryString(xxlbHexString);
        log.info("信息类别:"+xxlbBinString);

        String bl = xxlbBinString.substring(0,1);
        log.info("  ↓保留:"+bl);

        String bwtx = xxlbBinString.substring(1,3);
        log.info("  ↓报文通信:"+bwtx);

        String rimainData = xxlbBinString.substring(3,xxcdInt);
        byte txlx = Byte.parseByte(xxlbBinString.substring(3,8),2);
        log.info("通信类别:"+txlx);

        List<CommMsgSegModel> commMsgSegModelList = parsers.stream().filter(parser -> parser.support(txlx))
                .map(parser -> parser.parse(rimainData)).collect(Collectors.toList());
        if (commMsgSegModelList.size()>0&&commMsgSegModelList.get(0)!=null){
            CommMsgSegModel commMsgSegModel = commMsgSegModelList.get(0);
            Optional.ofNullable(commMsgSegModel).ifPresent(xfType004::setCommMsgSegModel);
        }  else{
            log.warn("还没解！！！！！！！！！");
            log.warn("余下数据:"+ rimainData);
        }

        return xfType004;
    }

    @Override
    public Boolean support(int messageTypeNo) {
        return messageTypeNo == 4;
    }


}
