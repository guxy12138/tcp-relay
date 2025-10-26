package org.example.models.xftype;

import lombok.Data;
import org.example.models.MessageDataModel;

@Data
public class XFType006 extends MessageDataModel {
    //入站网管标识-8bit
    private String inBoundFlag="00000001";
    //发方地址-24bit
    private String senderAddress;
    //信息长度-16bit
    private String messageLength;
    //保留字段-1bit
    private String remain;
    //网管信息
    private String networkManagement="10";
    //工作类别-5bit
    private String jobCategory;
}
