package org.example.models.xftype;

import lombok.Data;
import org.example.models.MessageDataModel;

//集团客户接入信息处理结果
@Data
public class XFType054 extends MessageDataModel {
    //信息编号-32bit
    private String messageNo;
    //发方类型-8bit
    private String senderType;
    //发信地址-48bit
    private String senderAddress;
    //收信地址-24bit
    private String receiverAddress;
    //时间-8bit
    private String hours;
    //分钟-8bit
    private String minutes;
    //秒-8bit
    private String seconds;
    //处理结果-8bit
    private String processResults;
    //未出站原因-8bit
    private String unOutBoundReasons;
}
