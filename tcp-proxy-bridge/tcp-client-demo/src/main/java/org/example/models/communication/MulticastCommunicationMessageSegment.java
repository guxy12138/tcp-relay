package org.example.models.communication;

import lombok.Data;
import org.example.models.CommMsgSegModel;

@Data
public class MulticastCommunicationMessageSegment extends CommMsgSegModel {
    //组播通讯
    private byte communType = 0b11000;
    //是否首帧
    private String ifFirstFrame;
    //有无续帧
    private String ifContinueFrame;
    //发方地址
    private String senderAddress;
    //发信时间
    private String senderTime;
    //辅助信息
    private String assertInfo;
    //编码类别
    private String codingType;
    //通信数据
    private String messageData;
}
