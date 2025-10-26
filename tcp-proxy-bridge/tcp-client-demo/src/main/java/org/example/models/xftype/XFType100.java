package org.example.models.xftype;

import lombok.Data;
import org.example.models.MessageDataModel;

import java.util.Optional;

@Data
public class XFType100 extends MessageDataModel {
    //32*8个字节
    String token;
    public String toString(){
        StringBuilder stringBuilder = new StringBuilder();
        if(Optional.ofNullable(token).isPresent()){
            stringBuilder.append(token);
            return stringBuilder.toString();
        }
        return null;
    }

    @Override
    public int sumLength() {
        return this.toString().length();
    }
}
