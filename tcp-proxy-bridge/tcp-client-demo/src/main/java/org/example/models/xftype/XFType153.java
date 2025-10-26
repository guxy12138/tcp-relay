package org.example.models.xftype;

import lombok.Data;
import org.example.models.MessageDataModel;

@Data
//集团客户接入传播明文报文信息
public class XFType153 extends MessageDataModel {
    //信息编号
    private String informationNo;
    //发方类型
    private String senderType;
    //发方地址
    private String senderAddress;
    //收方类型
    private String receiverType;
    //收方地址
    private String receiverAddress;
    //出站链路方式-8bit
    private String outboundLinkMode;
    //出站链路支持信息-24bit
    private String outboundLinkSupport;
    //信息长度-16bit
    private String messageLength;
    //信息类别-8bit
    private String messageType;
    //子类别-8bit
    private String subcategory;
    //信息内容
    private String informationContent;
}
