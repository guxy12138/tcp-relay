package org.example.models.message;

import lombok.Data;

@Data
public class DepartureMessageSegment {
    //发方地址-24bit
    private String senderAddress;
    //指令类型-4bit
    private String instructType;
    //指令代码(待定)
    private String instructCode;

    @Override
    public String toString() {
        return senderAddress + instructType + instructCode;
    }
}
