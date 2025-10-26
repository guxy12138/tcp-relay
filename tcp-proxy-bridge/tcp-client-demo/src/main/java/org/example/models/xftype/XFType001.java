package org.example.models.xftype;

import lombok.Data;
import org.example.models.MessageDataModel;
import org.example.models.message.GroundNetworkMessageSegment;
import org.example.util.BinaryTransUtils;

import java.util.Optional;

//XFType001:应急搜救请求服务信息
@Data
public class XFType001 extends MessageDataModel {
    //出站类型标识-8bit
    private String departure;
    //发方地址-24bit
    private String senderAddress;
    //发方入站响应卫星-8bit
    private String senderResSatellite;
    //发方入站波束号-8bit
    private String senderStackBeam;
    //收方类型-8bit
    private String receiverType;
    //收方地址-48bit
    private String receiverAddress;
    //信息长度-16bit
    private Integer messageLength;
    //保留字段-1bit
    private Integer remain;
    //出站帧类别-2bit
    private String stackOutType;
    //紧急标识-2bit
    private String emergencyFlag;
    //报告模式-3bit
    private String reportModel;
    private GroundNetworkMessageSegment groundNetworkMessageSegment;

    @Override
    public String toString() {
        StringBuilder stringBuilder = new StringBuilder();
        Optional.ofNullable(departure).ifPresent(stringBuilder::append);
        Optional.ofNullable(senderAddress).ifPresent(stringBuilder::append);
        Optional.ofNullable(senderResSatellite).ifPresent(stringBuilder::append);
        Optional.ofNullable(senderStackBeam).ifPresent(stringBuilder::append);
        Optional.ofNullable(receiverType).ifPresent(stringBuilder::append);
        Optional.ofNullable(receiverAddress).ifPresent(stringBuilder::append);
        Optional.ofNullable(BinaryTransUtils.intToBinaryString(messageLength,16)).ifPresent(stringBuilder::append);
        Optional.of(Integer.toBinaryString(remain)).ifPresent(stringBuilder::append);
        Optional.of(stackOutType).ifPresent(stringBuilder::append);
        Optional.of(emergencyFlag).ifPresent(stringBuilder::append);
        Optional.of(reportModel).ifPresent(stringBuilder::append);
        Optional.of(groundNetworkMessageSegment).ifPresent(stringBuilder::append);
        return stringBuilder.toString();
    }

    @Override
    public int sumLength() {
        return this.toString().length();
    }
}
