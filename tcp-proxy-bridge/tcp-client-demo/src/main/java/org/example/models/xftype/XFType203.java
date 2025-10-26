package org.example.models.xftype;

import lombok.Data;
import org.example.util.BinaryTransUtils;

@Data
public class XFType203 {
    //心跳信息
    private String heartbeat = BinaryTransUtils.hexStrToBinaryStr("E5BF83E8B7B3");
    private String separatorCharacter = BinaryTransUtils.hexStrToBinaryStr("7878787888888888");

    public String toString() {
        return heartbeat + separatorCharacter;
    }
}