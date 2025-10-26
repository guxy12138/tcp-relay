package org.example.models;

import com.alibaba.fastjson2.annotation.JSONField;
import lombok.Data;
import org.example.util.BinaryTransUtils;

@Data
public class DataModel {
    //信息类型编号 -2Byte
    @JSONField(name = "信息类型编号")
    private Integer messageTypeNo;
    //信息内容长度 -2Byte
    @JSONField(name = "信息内容长度")
    private Integer messageLength;
    //信息内容
    @JSONField(name = "信息内容")
    private MessageDataModel messageData;

    @Override
    public String toString() {
        return BinaryTransUtils.intToBinaryString(messageTypeNo, 16)
                + BinaryTransUtils.intToBinaryString(messageLength, 16)
                + messageData;
    }
}
