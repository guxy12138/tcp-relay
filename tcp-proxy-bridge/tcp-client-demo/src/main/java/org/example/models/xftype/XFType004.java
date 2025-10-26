package org.example.models.xftype;

import com.alibaba.fastjson2.annotation.JSONField;
import lombok.Data;
import org.example.models.CommMsgSegModel;
import org.example.models.MessageDataModel;

@Data
public class XFType004 extends MessageDataModel {
    //出站类型标识-8bit
    @JSONField(name="出站类型标识")
    private Integer departure;
    //收方类型-8bit
    @JSONField(name="收方类型")
    private Integer recieverType;
    //收方地址-48bit
    @JSONField(name="收方地址")
    private Long recieverAddress;
    //信息长度-16bit
    @JSONField(name="信息长度")
    private Integer messageLength;
    //保留-1bit
    @JSONField(name="保留")
    private Integer remain;
    //报文通信
    @JSONField(name="报文通信")
    private Integer messageCommunication;
    //通信类型
    @JSONField(name="通信类型")
    private String communicationType;
    //出站信息段
    @JSONField(name="出站信息段")
    private CommMsgSegModel commMsgSegModel;
}
