package com.study.client;

import java.io.File;
import java.net.URI;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.TimeUnit;
import java.util.logging.Logger;

import javax.net.ssl.SSLException;

import com.study.etcd.EtcdNameResolverProvider;
import com.study.pb.GreeterGrpc;
import com.study.pb.HelloReply;
import com.study.pb.HelloRequest;

import io.grpc.ManagedChannel;
import io.grpc.StatusRuntimeException;
import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NegotiationType;
import io.grpc.netty.NettyChannelBuilder;
import io.netty.handler.ssl.SslContext;
import io.netty.handler.ssl.SslContextBuilder;

public class HelloWorldClientTLS {
	private static final Logger logger = Logger.getLogger(HelloWorldClientTLS.class.getName());
	private static final String ENDPOINT = "http://127.0.0.1:2379";
	private static final String TARGET = "etcd:///grpclb";

	private final ManagedChannel channel;
	private final GreeterGrpc.GreeterBlockingStub blockingStub;

	private static SslContext buildSslContext(String trustCertCollectionFilePath, String clientCertChainFilePath,
			String clientPrivateKeyFilePath) throws SSLException {
		SslContextBuilder builder = GrpcSslContexts.forClient();
		if (trustCertCollectionFilePath != null) {
			builder.trustManager(new File(trustCertCollectionFilePath));
		}
		if (clientCertChainFilePath != null && clientPrivateKeyFilePath != null) {
			builder.keyManager(new File(clientCertChainFilePath), new File(clientPrivateKeyFilePath));
		}
		return builder.build();
	}

	public HelloWorldClientTLS(SslContext sslContext) throws SSLException {
//		NettyChannelBuilder.

		List<URI> endpoints = new ArrayList<>();
		endpoints.add(URI.create(ENDPOINT));
		this.channel = NettyChannelBuilder.forTarget(TARGET)
				.nameResolverFactory(EtcdNameResolverProvider.forEndpoints(endpoints))
				.defaultLoadBalancingPolicy("round_robin")
				.negotiationType(NegotiationType.TLS).sslContext(sslContext)
//				.usePlaintext()
				.build();
		blockingStub = GreeterGrpc.newBlockingStub(channel);
	}

	HelloWorldClientTLS(ManagedChannel channel) {
		this.channel = channel;
		blockingStub = GreeterGrpc.newBlockingStub(channel);
	}

	public void shutdown() throws InterruptedException {
		channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
	}

	public void greet(String name) {
		HelloRequest request = HelloRequest.newBuilder().setName(name).build();
		HelloReply response;
		try {
			response = blockingStub.sayHello(request);
		} catch (StatusRuntimeException e) {
			e.printStackTrace();
			return;
		}
		System.out.println(response.getMessage());
	}

	public static void main(String[] args) throws Exception {
		String path = HelloWorldClientTLS.class.getResource("/").toString();
		path = path.replace("file:", "");
		path = "/home/shepard/tmp/";
		String caCertFile = path + "/certs/ca.crt";
		String certFile = path + "/certs/client.crt";
		String keyFile = path + "/certs/client.pem";
		HelloWorldClientTLS client = null;
		try {
			client = new HelloWorldClientTLS(buildSslContext(caCertFile, certFile, keyFile));
			System.out.println(client == null);
			client.greet("tom");
		} finally {
			client.shutdown();
		}
	}
}
