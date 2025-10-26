package org.example.models.message;

import lombok.Data;

import java.util.Optional;

//地面网出信息段
@Data
public class GroundNetworkMessageSegment {
    //发方地址-24bit
    private String senderAddress;
    //位置报告数据-93bits
    private String locationReportData;
    //搜救业务子类型
    private String searchRescueType;
    //搜救业务数据
    private String searchRescueData;

    @Override
    public String toString() {
        StringBuilder stringBuilder = new StringBuilder();
        Optional.ofNullable(senderAddress).ifPresent(stringBuilder::append);
        Optional.ofNullable(locationReportData).ifPresent(stringBuilder::append);
        Optional.ofNullable(searchRescueType).ifPresent(stringBuilder::append);
        Optional.ofNullable(searchRescueData).ifPresent(stringBuilder::append);
        return stringBuilder.toString();
    }
}
