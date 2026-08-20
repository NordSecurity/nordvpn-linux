extension StringTranslation on String {
  int commonPrefixLength(String other) {
    final length = this.length < other.length ? this.length : other.length;

    var i = 0;
    while (i < length && this[i] == other[i]) {
      i++;
    }

    return i;
  }
}
