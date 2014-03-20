require 'formula'

class Slang < Formula
  
  url 'https://github.com/bww/Slang.git'
  homepage 'https://github.com/bww/Slang'
  version '1'
  
  depends_on 'go'
  
  def install
    # build Slang
    system "make"
    # install Slang
    bin.install("slang")
    # install resources here...
    share.install("resources")
  end
  
end

