package org.example;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class Application {
    public static void main(String[] args) {
        final Logger logger = LogManager.getLogger();
        try {
            SpringApplication.run(Application.class, args);
        }catch (Exception e){
           logger.error(e);
        }
    }
}