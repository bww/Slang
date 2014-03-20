require 'formula'

class Slang < Formula
  
  env :std # this is necessary for CGO, apparently
  
  url 'https://github.com/bww/Slang.git'
  homepage 'https://github.com/bww/Slang'
  version '1'
  
  depends_on 'go'
  
  def install
    # build Slang
    system "make"
    # install Slang
    bin.install("bin/slang")
    # install resources here...
    share.install("share/slang")
  end
  
end

