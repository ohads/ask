class Ask < Formula
  desc "A CLI tool to get ChatGPT answers from the command line"
  homepage "https://github.com/yourusername/ask"
  version "1.0.0"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/yourusername/ask/releases/download/v1.0.0/ask-darwin-arm64"
    sha256 "PLACEHOLDER_SHA256"
  elsif OS.mac? && Hardware::CPU.intel?
    url "https://github.com/yourusername/ask/releases/download/v1.0.0/ask-darwin-amd64"
    sha256 "PLACEHOLDER_SHA256"
  elsif OS.linux? && Hardware::CPU.arm?
    url "https://github.com/yourusername/ask/releases/download/v1.0.0/ask-linux-arm64"
    sha256 "PLACEHOLDER_SHA256"
  elsif OS.linux? && Hardware::CPU.intel?
    url "https://github.com/yourusername/ask/releases/download/v1.0.0/ask-linux-amd64"
    sha256 "PLACEHOLDER_SHA256"
  end

  def install
    if OS.mac? && Hardware::CPU.arm?
      bin.install "ask-darwin-arm64" => "ask"
    elsif OS.mac? && Hardware::CPU.intel?
      bin.install "ask-darwin-amd64" => "ask"
    elsif OS.linux? && Hardware::CPU.arm?
      bin.install "ask-linux-arm64" => "ask"
    elsif OS.linux? && Hardware::CPU.intel?
      bin.install "ask-linux-amd64" => "ask"
    end
  end

  test do
    system "#{bin}/ask", "--help"
  end
end 