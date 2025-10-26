package org.example.models.communication;

import lombok.Data;
import org.example.models.CommMsgSegModel;

@Data
public class ReceiptQueryResult extends CommMsgSegModel {
    //回执查询结果
    private byte communType= 0b100000;
    //回执查询结果
    private String results;
}
