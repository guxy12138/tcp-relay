package org.example.models.xftype;

import lombok.Data;
import org.example.models.MessageDataModel;
import org.example.models.message.DepartureMessageSegment;

//搜救中心指示代码
@Data
public class XFType002 extends MessageDataModel {
    //出站类型标识-8bit
    private String departure;
    //收方类型-8bit
    private String recieverType;
    //收方地址-48bit
    private String recieverAddress;
    //信息长度-16bit
    private Integer messageLen;
    //保留-1bit
    private String remain;
    //出站帧类别
    private String departureFrame;
    //紧急标识-2bit
    private String emergencyFlag;
    //报告模式-3bit
    private String reportModel;
    //出站信息段
    private DepartureMessageSegment departureMessageSegment;
}
