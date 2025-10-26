package org.example.compent.parser;

import org.example.models.MessageDataModel;

public interface XFTypeParser {
    MessageDataModel parse(String data) throws Exception;

    Boolean support(int messageTypeNo);
}
