package org.example.util.message.xftype100;

import org.example.Application;
import org.example.models.BasePackageModel;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.ApplicationRunner;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

import java.security.NoSuchAlgorithmException;

import static org.junit.Assert.*;
@SpringBootTest(classes = Application.class)
@RunWith(SpringRunner.class)
public class BaseInitXFtype100Test {
    @Autowired
    private BaseInitXFtype100 baseInitXFtype100;

    @Test
    public void initXFType100() throws NoSuchAlgorithmException {
        BasePackageModel basePackageModel=baseInitXFtype100.initXFType100();
    }
}