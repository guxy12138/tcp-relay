package org.example.models.xftype;

import lombok.Data;
import org.example.models.MessageDataModel;
import org.example.models.message.PosDepartureMessageSegment;

@Data
//位置报告服务信息
public class XFType003 extends MessageDataModel {
    //出站类型标识-8bit
    private String departure;
    //收方类型-8bit
    private String recieverType;
    //收方地址-48bit
    private String recieverAddress;
    //信息长度-16bit
    private Integer messageLength;
    //保留-1bit
    private String remain;
    //出站帧类别
    private String departureFrame;
    //紧急标识-2bit
    private String emergency;
    //报告模式-3bit
    private String reportModel;
    //出站信息段
    private PosDepartureMessageSegment posDepartureMessageSegment;
}
