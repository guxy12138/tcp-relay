package org.example.models.broadcast;

import lombok.Data;

@Data
public class Position {
    //经度-9bit
    private String longitude;
    //纬度-8bit
    private String latitude;
}
