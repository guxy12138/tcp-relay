package org.example.models;

import com.alibaba.fastjson2.annotation.JSONField;
import lombok.Data;

import java.util.Optional;

@Data
public class DataSegmentModels {
    //当前数据段
    @JSONField(name="当前数据段")
    private DataSegmentModel currentDataSegment;
    //重发数据段
    @JSONField(name="重发数据段")
    private DataSegmentModel retransmissionDataSegment;

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder();
        Optional.ofNullable(currentDataSegment).ifPresent(sb::append);
        Optional.ofNullable(retransmissionDataSegment).ifPresent(sb::append);
        return sb.toString();
    }
}
