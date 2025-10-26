package org.example.models.communication;

import com.alibaba.fastjson2.annotation.JSONField;
import lombok.Data;
import org.example.models.CommMsgSegModel;

import java.util.Date;

@Data
public class BunchPlantCommunicationMessageSegment extends CommMsgSegModel {
    //点播通讯
    @JSONField(name = "通讯类型")
    private byte communType = 0b10;
    //发方类型-1bit
    @JSONField(name = "发方类型")
    private byte senderType;
    //实时标识-1bit
    @JSONField(name = "实时标识")
    private byte actualTime;
    //是否首帧
    @JSONField(name = "是否首帧")
    private byte ifFirstFrame;
    //有无连续帧
    @JSONField(name = "有无连续帧")
    private byte ifConsecutiveFrame;
    //发方地址
    @JSONField(name = "发方地址")
    private Long senderAddress;
    //发信时间
    @JSONField(name = "发信时间")
    private Date senderTime;
    //编码类型
    @JSONField(name = "编码类型")
    private byte codingType;
    //通信数据
    @JSONField(name = "通信数据")
    private String messageData;
}
