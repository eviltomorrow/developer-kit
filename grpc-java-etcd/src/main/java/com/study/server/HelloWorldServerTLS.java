package com.study.server;

import java.io.File;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.URI;
import java.util.concurrent.ExecutionException;
import java.util.logging.Logger;

import com.google.common.base.Charsets;
import com.study.pb.GreeterGrpc.GreeterImplBase;
import com.study.pb.HelloReply;
import com.study.pb.HelloRequest;

import io.etcd.jetcd.ByteSequence;
import io.etcd.jetcd.Client;
import io.etcd.jetcd.lease.LeaseKeepAliveResponse;
import io.etcd.jetcd.options.PutOption;
import io.grpc.Server;
import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NettyServerBuilder;
import io.grpc.stub.StreamObserver;
import io.netty.handler.ssl.ClientAuth;
import io.netty.handler.ssl.SslContextBuilder;
import io.netty.handler.ssl.SslProvider;

public class HelloWorldServerTLS {
	private static final Logger logger = Logger.getLogger(HelloWorldServerTLS.class.getName());
	private Server server;

	private final String host;
	private final int port;
	private final String certChainFilePath;
	private final String privateKeyFilePath;
	private final String trustCertCollectionFilePath;

	private static final String ENDPOINT = "http://127.0.0.1:2379";
	private static final String target = "grpclb/";
	private static final long TTL = 5L;
	private Client etcd;

	public HelloWorldServerTLS(String host, int port, String certChainFilePath, String privateKeyFilePath,
			String trustCertCollectionFilePath) {
		this.host = host;
		this.port = port;
		this.certChainFilePath = certChainFilePath;
		this.privateKeyFilePath = privateKeyFilePath;
		this.trustCertCollectionFilePath = trustCertCollectionFilePath;

	}

	private SslContextBuilder getSslContextBuilder() {
		SslContextBuilder sslClientContextBuilder = SslContextBuilder.forServer(new File(certChainFilePath),
				new File(privateKeyFilePath));
		if (trustCertCollectionFilePath != null) {
			sslClientContextBuilder.trustManager(new File(trustCertCollectionFilePath));
			sslClientContextBuilder.clientAuth(ClientAuth.REQUIRE);
		}
		return GrpcSslContexts.configure(sslClientContextBuilder, SslProvider.OPENSSL);
	}

	private void start() throws IOException, InterruptedException, ExecutionException {
		server = NettyServerBuilder.forAddress(new InetSocketAddress(host, port)).addService(new GreeterImpl())
				.sslContext(getSslContextBuilder().build())
				.build().start();
		logger.info("Server started on port:" + port);

		final URI uri = URI.create("http://localhost:" + port);
		this.etcd = Client.builder().endpoints(URI.create(ENDPOINT)).build();
		long leaseId = etcd.getLeaseClient().grant(TTL).get().getID();
		etcd.getKVClient().put(ByteSequence.from(target + uri.toASCIIString(), Charsets.US_ASCII),
				ByteSequence.from(Long.toString(leaseId), Charsets.US_ASCII),
				PutOption.newBuilder().withLeaseId(leaseId).build());
		etcd.getLeaseClient().keepAlive(leaseId, new EtcdServiceRegisterer());

		Runtime.getRuntime().addShutdownHook(new Thread() {
			@Override
			public void run() {
				System.err.println("*** shutting down gRPC server since JVM is shutting down");
				HelloWorldServerTLS.this.stop();
				System.err.println("*** server shut down");
			}
		});
	}

	private void stop() {
		if (server != null) {
			server.shutdown();
		}
	}

	private void blockUntilShutdown() throws InterruptedException {
		if (server != null) {
			server.awaitTermination();
		}
	}

	public static void main(String[] args) throws IOException, InterruptedException, ExecutionException {
		String path = HelloWorldServerTLS.class.getResource("/").toString();
		path = path.replace("file:", "");
		path = "/home/shepard/tmp";
		String host = "localhost";
		int port = 8080;
		String certFile = path + "/certs/server.crt";
		String keyFile = path + "/certs/server.pem";
		String caCertFile = path + "/certs/ca.crt";

		final HelloWorldServerTLS server = new HelloWorldServerTLS(host, port, certFile, keyFile, caCertFile);
		server.start();
		server.blockUntilShutdown();
	}

	static class GreeterImpl extends GreeterImplBase {
		public void sayHello(HelloRequest req, StreamObserver<HelloReply> responseObserver) {
			HelloReply reply = HelloReply.newBuilder().setMessage("Hello " + req.getName()).build();
			System.out.println(req.getName());
			responseObserver.onNext(reply);
			responseObserver.onCompleted();
		}
	}

	class EtcdServiceRegisterer implements StreamObserver<LeaseKeepAliveResponse> {
		@Override
		public void onNext(LeaseKeepAliveResponse value) {
//			logger.info("got renewal for lease: " + value.getID());
		}

		@Override
		public void onError(Throwable t) {
		}

		@Override
		public void onCompleted() {
		}
	}
}
