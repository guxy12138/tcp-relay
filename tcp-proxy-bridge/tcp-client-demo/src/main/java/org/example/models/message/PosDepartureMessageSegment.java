package org.example.models.message;

import lombok.Data;

@Data
public class PosDepartureMessageSegment {
    //发送方地址-24bits
    private String senderAddress;
    //位置报告数据-93bit
    private String positionReportData;
    //状态数据-636bit
    private String statusData;

    @Override
    public String toString() {
        return senderAddress + positionReportData + statusData;
    }
}
