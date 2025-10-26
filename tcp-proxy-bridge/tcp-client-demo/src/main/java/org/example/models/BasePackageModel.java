package org.example.models;

import com.alibaba.fastjson2.annotation.JSONField;
import lombok.Data;
import org.example.util.BinaryTransUtils;

@Data
public class BasePackageModel {
    //信源 - 2Byte -16bit
    @JSONField(name = "信源")
    private Integer sourceInfo;
    //信宿 -2Byte
    @JSONField(name = "信宿")
    private Integer hostInfo;
    //包序号 -4Byte(32位)
    @JSONField(name = "包序号")
    private Long packageNo;
    //当前数据项 -1Byte
    @JSONField(name = "当前数据项")
    private Integer currentDataItem;
    //当前数据总长度 -2Byte
    @JSONField(name = "当前数据总长度")
    private Integer dataSumLength;
    //重发标志 -1Byte
    @JSONField(name = "重发标志")
    private Integer retransmissionFlag;
    //重发数据项 -1Byte
    @JSONField(name = "重发数据项")
    private Integer retransmissionData;
    //重发数据总长度 -2Byte
    @JSONField(name = "重发数据总长度")
    private Integer retransmissionSumLength;
    //数据段
    @JSONField(name = "数据段")
    private DataSegmentModels dataSegments;
    //TCP粘包分隔符
    @JSONField(name = "Tcp粘包分隔符")
    private String separatorCharacter = BinaryTransUtils.hexStrToBinaryStr("7878787888888888");

    @Override
    public String toString() {
        return BinaryTransUtils.intToBinaryString(sourceInfo, 16)
                + BinaryTransUtils.intToBinaryString(hostInfo, 16)
                + BinaryTransUtils.longToBinaryString(packageNo,32)
                + BinaryTransUtils.intToBinaryString(currentDataItem, 8)
                + BinaryTransUtils.intToBinaryString(dataSumLength, 16)
                + BinaryTransUtils.intToBinaryString(retransmissionFlag, 8)
                + BinaryTransUtils.intToBinaryString(retransmissionData, 8)
                + BinaryTransUtils.intToBinaryString(retransmissionSumLength, 16)
                + dataSegments
                + separatorCharacter;
    }
}
