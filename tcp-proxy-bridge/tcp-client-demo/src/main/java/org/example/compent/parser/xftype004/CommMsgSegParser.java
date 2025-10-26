package org.example.compent.parser.xftype004;

import org.example.models.CommMsgSegModel;

/**
 * @Description
 * @Author suoshen
 * @Date 2023/5/18 17:01
 * @Version 1.0
 */
public interface CommMsgSegParser {

    CommMsgSegModel parse(String briData);

    Boolean support(byte commType);
}
