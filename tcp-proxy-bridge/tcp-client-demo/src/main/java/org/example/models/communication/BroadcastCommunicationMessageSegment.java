package org.example.models.communication;

import lombok.Data;
import org.example.models.CommMsgSegModel;

@Data
public class BroadcastCommunicationMessageSegment extends CommMsgSegModel {
    //通播通讯
    private byte communType = 0b11010;
    //是否首帧
    private String ifFirstFrame;
    //有无连续帧
    private String ifConsecutiveFrame;
    //发方地址
    private String senderAddress;
    //发信时间
    private String senderTime;
    //编码类型
    private String codingType;
    //通信数据
    private String messageData;
}
