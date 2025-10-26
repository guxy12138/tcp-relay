package org.example.models.xftype;

import lombok.Data;
import org.example.models.MessageDataModel;
import org.example.models.broadcast.FixedDomain;
import org.example.models.broadcast.MessageType;

@Data
public class XFType005 extends MessageDataModel {
    //固定域
    private FixedDomain fixedDomain;
    //MsgType
    private MessageType messageType;
    //类别信息域
    private String categoryInfo;
}
