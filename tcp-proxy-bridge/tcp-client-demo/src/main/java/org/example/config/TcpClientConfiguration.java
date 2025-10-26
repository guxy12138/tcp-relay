package org.example.config;

import lombok.Data;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Configuration;

@Configuration
@Data
public class TcpClientConfiguration {
    @Value("${client.address}")
    private String address;
    @Value("${client.port}")
    private int port;
}
