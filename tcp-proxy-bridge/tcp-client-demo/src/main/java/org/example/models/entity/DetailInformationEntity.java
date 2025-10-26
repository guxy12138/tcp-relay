package org.example.models.entity;

import lombok.Data;
import org.springframework.format.annotation.DateTimeFormat;

import java.io.Serializable;
import java.util.Date;


@Data
public class DetailInformationEntity implements Serializable {

    private Integer id;
    //发方地址
    private String senderAddress;
    //收方地址
    private String recipientAddress;
    //入库时间
    @DateTimeFormat(pattern="yyyy-MM-dd HH:mm:ss")
    private Date date;
    //详细信息
    private String details;
    //预留字段
    private String reserve1;
}
