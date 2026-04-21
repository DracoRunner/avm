class Avm < Formula
  desc "Lightweight local/global command alias manager (like asdf/nvm for aliases)"
  homepage "https://github.com/DracoRunner/avm"
  version "0.2.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DracoRunner/avm/releases/download/v0.2.0/avm_darwin_arm64.tar.gz"
      sha256 "bf13a6e7759551d1cdc0396d8029bbacdee871cfc5a5bc671d6c7f1d488b444e"
    else
      url "https://github.com/DracoRunner/avm/releases/download/v0.2.0/avm_darwin_amd64.tar.gz"
      sha256 "c7e09b94f22ddc70afbe41588994ca1146d0b4aaa42e5cc1d79a86d4041a1ff2"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DracoRunner/avm/releases/download/v0.2.0/avm_linux_arm64.tar.gz"
      sha256 "6cb614618db6fb5f61454d27c766db9b41f01a1ee19e9e6a7ab9d65c7d9240be"
    else
      url "https://github.com/DracoRunner/avm/releases/download/v0.2.0/avm_linux_amd64.tar.gz"
      sha256 "41755aeb8aa4f114b5266eb7a6f4b7c2cdd66349a2661408934d79c486327f0b"
    end
  end

  def install
    bin.install "avm" => "avm-bin"
  end

  def caveats
    <<~EOS
      To enable avm in your shell, add this to your ~/.zshrc or ~/.bashrc:

        eval "$(avm-bin shell-init)"

      Then reload your shell:

        source ~/.zshrc  # or source ~/.bashrc
    EOS
  end

  test do
    system "#{bin}/avm-bin", "version"
  end
end
